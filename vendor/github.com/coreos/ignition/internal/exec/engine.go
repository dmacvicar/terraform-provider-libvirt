// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exec

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"time"

	"github.com/coreos/ignition/config"
	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/config/validate/report"
	"github.com/coreos/ignition/internal/exec/stages"
	"github.com/coreos/ignition/internal/exec/util"
	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/oem"
	"github.com/coreos/ignition/internal/providers"
	"github.com/coreos/ignition/internal/providers/cmdline"
	"github.com/coreos/ignition/internal/resource"
	internalUtil "github.com/coreos/ignition/internal/util"
)

const (
	DefaultFetchTimeout = time.Minute
)

// Engine represents the entity that fetches and executes a configuration.
type Engine struct {
	ConfigCache  string
	FetchTimeout time.Duration
	Logger       *log.Logger
	Root         string
	OEMConfig    oem.Config

	client resource.HttpClient
}

// Run executes the stage of the given name. It returns true if the stage
// successfully ran and false if there were any errors.
func (e Engine) Run(stageName string) bool {
	cfg, f, err := e.acquireConfig()
	switch err {
	case nil:
	case config.ErrCloudConfig, config.ErrScript, config.ErrEmpty:
		e.Logger.Info("%v: ignoring user-provided config", err)
		cfg = e.OEMConfig.DefaultUserConfig()
	default:
		e.Logger.Crit("failed to acquire config: %v", err)
		return false
	}

	e.Logger.PushPrefix(stageName)
	defer e.Logger.PopPrefix()

	baseConfig := types.Config{
		Ignition: types.Ignition{Version: types.MaxVersion.String()},
		Storage: types.Storage{
			Filesystems: []types.Filesystem{{
				Name: "root",
				Path: internalUtil.StringToPtr(e.Root),
			}},
		},
	}

	return stages.Get(stageName).Create(e.Logger, e.Root, f).Run(config.Append(baseConfig, config.Append(e.OEMConfig.BaseConfig(), cfg)))
}

// acquireConfig returns the configuration, first checking a local cache
// before attempting to fetch it from the provider.
func (e *Engine) acquireConfig() (cfg types.Config, f resource.Fetcher, err error) {
	// First try read the config @ e.ConfigCache.
	b, err := ioutil.ReadFile(e.ConfigCache)
	if err == nil {
		if err = json.Unmarshal(b, &cfg); err != nil {
			e.Logger.Crit("failed to parse cached config: %v", err)
		}
		// Create an http client and fetcher with the timeouts from the cached
		// config
		e.client = resource.NewHttpClient(e.Logger, cfg.Ignition.Timeouts)
		f, err = e.OEMConfig.NewFetcherFunc()(e.Logger, &e.client)
		if err != nil {
			e.Logger.Crit("failed to generate fetcher: %s", err)
			return
		}
		return
	}

	// Create a new http client and fetcher with the timeouts set via the flags,
	// since we don't have a config with timeout values we can use
	timeout := int(e.FetchTimeout.Seconds())
	e.client = resource.NewHttpClient(e.Logger, types.Timeouts{HTTPTotal: &timeout})
	f, err = e.OEMConfig.NewFetcherFunc()(e.Logger, &e.client)
	if err != nil {
		e.Logger.Crit("failed to generate fetcher: %s", err)
		return
	}

	// (Re)Fetch the config if the cache is unreadable.
	cfg, err = e.fetchProviderConfig(f)
	if err != nil {
		e.Logger.Crit("failed to fetch config: %s", err)
		return
	}

	// Regenerate the http client and fetcher to use the timeouts from the
	// newly fetched config
	e.client = resource.NewHttpClient(e.Logger, cfg.Ignition.Timeouts)
	f, err = e.OEMConfig.NewFetcherFunc()(e.Logger, &e.client)
	if err != nil {
		e.Logger.Crit("failed to generate fetcher: %s", err)
		return
	}

	// Populate the config cache.
	b, err = json.Marshal(cfg)
	if err != nil {
		e.Logger.Crit("failed to marshal cached config: %v", err)
		return
	}
	if err = ioutil.WriteFile(e.ConfigCache, b, 0640); err != nil {
		e.Logger.Crit("failed to write cached config: %v", err)
		return
	}

	return
}

// fetchProviderConfig returns the externally-provided configuration. It first
// checks to see if the command-line option is present. If so, it uses that
// source for the configuration. If the command-line option is not present, it
// check's the engine's provider. An error is returned if the provider is
// unavailable. This will also render the config (see renderConfig) before
// returning.
func (e Engine) fetchProviderConfig(f resource.Fetcher) (types.Config, error) {
	cfg, r, err := cmdline.FetchConfig(f)
	if err == providers.ErrNoProvider {
		cfg, r, err = e.OEMConfig.FetchFunc()(f)
	}

	e.logReport(r)
	if err != nil {
		return types.Config{}, err
	}

	return e.renderConfig(cfg, f)
}

// renderConfig evaluates "ignition.config.replace" and "ignition.config.append"
// in the given config and returns the result. If "ignition.config.replace" is
// set, the referenced and evaluted config will be returned. Otherwise, if
// "ignition.config.append" is set, each of the referenced configs will be
// evaluated and appended to the provided config. If neither option is set, the
// provided config will be returned unmodified.
func (e *Engine) renderConfig(cfg types.Config, f resource.Fetcher) (types.Config, error) {
	// Apply any new timeout info before fetching other configs.
	e.client = resource.NewHttpClient(e.Logger, cfg.Ignition.Timeouts)
	if cfgRef := cfg.Ignition.Config.Replace; cfgRef != nil {
		return e.fetchReferencedConfig(*cfgRef, f)
	}

	appendedCfg := cfg
	for _, cfgRef := range cfg.Ignition.Config.Append {
		newCfg, err := e.fetchReferencedConfig(cfgRef, f)
		if err != nil {
			return newCfg, err
		}

		appendedCfg = config.Append(appendedCfg, newCfg)
	}
	return appendedCfg, nil
}

// fetchReferencedConfig fetches, renders, and attempts to verify the requested
// config.
func (e Engine) fetchReferencedConfig(cfgRef types.ConfigReference, f resource.Fetcher) (types.Config, error) {
	u, err := url.Parse(cfgRef.Source)
	if err != nil {
		return types.Config{}, err
	}
	rawCfg, err := f.FetchToBuffer(*u, resource.FetchOptions{
		Headers: resource.ConfigHeaders,
	})
	if err != nil {
		return types.Config{}, err
	}

	if err := util.AssertValid(cfgRef.Verification, rawCfg); err != nil {
		return types.Config{}, err
	}

	cfg, r, err := config.Parse(rawCfg)
	e.logReport(r)
	if err != nil {
		return types.Config{}, err
	}

	return e.renderConfig(cfg, f)
}

func (e Engine) logReport(r report.Report) {
	for _, entry := range r.Entries {
		switch entry.Kind {
		case report.EntryError:
			e.Logger.Crit("%v", entry)
		case report.EntryWarning:
			e.Logger.Warning("%v", entry)
		case report.EntryDeprecated:
			e.Logger.Warning("%v: the provided config format is deprecated and will not be supported in the future.", entry)
		case report.EntryInfo:
			e.Logger.Info("%v", entry)
		}
	}
}
