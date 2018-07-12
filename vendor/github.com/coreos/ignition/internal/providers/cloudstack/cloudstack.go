// Copyright 2017 CoreOS, Inc.
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

// The CloudStack provider fetches configurations from the userdata available in
// the config-drive.
// NOTE: This provider is still EXPERIMENTAL.

package cloudstack

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/coreos/ignition/config/validate/report"
	"github.com/coreos/ignition/internal/config"
	"github.com/coreos/ignition/internal/config/types"
	"github.com/coreos/ignition/internal/distro"
	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/resource"
)

const (
	configDriveUserdataPath = "/cloudstack/userdata/user_data.txt"
	LeaseRetryInterval      = 500 * time.Millisecond
)

func FetchConfig(f resource.Fetcher) (types.Config, report.Report, error) {
	var data []byte
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	dispatch := func(name string, fn func() ([]byte, error)) {
		raw, err := fn()
		if err != nil {
			switch err {
			case context.Canceled:
			case context.DeadlineExceeded:
				f.Logger.Err("timed out while fetching config from %s", name)
			default:
				f.Logger.Err("failed to fetch config from %s: %v", name, err)
			}
			return
		}

		data = raw
		cancel()
	}

	go dispatch("config drive (config)", func() ([]byte, error) {
		return fetchConfigFromDevice(f.Logger, ctx, "config-2")
	})

	go dispatch("config drive (CONFIG)", func() ([]byte, error) {
		return fetchConfigFromDevice(f.Logger, ctx, "CONFIG-2")
	})

	go dispatch("metadata service", func() ([]byte, error) {
		return fetchConfigFromMetadataService(f)
	})

	<-ctx.Done()
	if ctx.Err() == context.DeadlineExceeded {
		f.Logger.Info("neither config drive nor metadata service were available in time. Continuing without a config...")
	}

	return config.Parse(data)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return (err == nil)
}

func labelExists(label string) bool {
	_, err := getPath(label)
	return (err == nil)
}

func getPath(label string) (string, error) {
	path := filepath.Join(distro.DiskByLabelDir(), label)

	if fileExists(path) {
		return path, nil
	}

	return "", fmt.Errorf("label not found: %s", label)
}

func findLease() (*os.File, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("could not list interfaces: %v", err)
	}

	for {
		for _, iface := range ifaces {
			lease, err := os.Open(fmt.Sprintf("/run/systemd/netif/leases/%d", iface.Index))
			if os.IsNotExist(err) {
				continue
			} else if err != nil {
				return nil, err
			} else {
				return lease, nil
			}
		}

		fmt.Printf("No leases found. Waiting...")
		time.Sleep(LeaseRetryInterval)
	}
}

func getDHCPServerAddress() (string, error) {
	lease, err := findLease()
	if err != nil {
		return "", err
	}
	defer lease.Close()

	var address string
	line := bufio.NewScanner(lease)
	for line.Scan() {
		parts := strings.Split(line.Text(), "=")
		if parts[0] == "SERVER_ADDRESS" && len(parts) == 2 {
			address = parts[1]
			break
		}
	}

	if len(address) == 0 {
		return "", fmt.Errorf("dhcp server address not found in leases")
	}

	return address, nil
}

func fetchConfigFromDevice(logger *log.Logger, ctx context.Context, label string) ([]byte, error) {
	for !labelExists(label) {
		logger.Debug("config drive (%q) not found. Waiting...", label)
		select {
		case <-time.After(time.Second):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	path, err := getPath(label)
	if err != nil {
		return nil, err
	}

	logger.Debug("creating temporary mount point")
	mnt, err := ioutil.TempDir("", "ignition-configdrive")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.Remove(mnt)

	cmd := exec.Command(distro.MountCmd(), "-o", "ro", "-t", "auto", path, mnt)
	if _, err := logger.LogCmd(cmd, "mounting config drive"); err != nil {
		return nil, err
	}
	defer logger.LogOp(
		func() error { return syscall.Unmount(mnt, 0) },
		"unmounting %q at %q", path, mnt,
	)

	if !fileExists(filepath.Join(mnt, configDriveUserdataPath)) {
		return nil, nil
	}

	return ioutil.ReadFile(filepath.Join(mnt, configDriveUserdataPath))
}

func fetchConfigFromMetadataService(f resource.Fetcher) ([]byte, error) {
	addr, err := getDHCPServerAddress()
	if err != nil {
		return nil, err
	}

	metadataServiceUrl := url.URL{
		Scheme: "http",
		Host:   addr,
		Path:   "/latest/user-data",
	}

	res, err := f.FetchToBuffer(metadataServiceUrl, resource.FetchOptions{
		Headers: resource.ConfigHeaders,
	})
	return res, err
}
