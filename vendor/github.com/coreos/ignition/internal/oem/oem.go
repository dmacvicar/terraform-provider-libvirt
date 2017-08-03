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

package oem

import (
	"fmt"
	"net/url"

	"github.com/coreos/ignition/config/types"
	"github.com/coreos/ignition/internal/log"
	"github.com/coreos/ignition/internal/providers"
	"github.com/coreos/ignition/internal/providers/azure"
	"github.com/coreos/ignition/internal/providers/digitalocean"
	"github.com/coreos/ignition/internal/providers/ec2"
	"github.com/coreos/ignition/internal/providers/file"
	"github.com/coreos/ignition/internal/providers/gce"
	"github.com/coreos/ignition/internal/providers/noop"
	"github.com/coreos/ignition/internal/providers/openstack"
	"github.com/coreos/ignition/internal/providers/packet"
	"github.com/coreos/ignition/internal/providers/qemu"
	"github.com/coreos/ignition/internal/providers/virtualbox"
	"github.com/coreos/ignition/internal/providers/vmware"
	"github.com/coreos/ignition/internal/registry"
	"github.com/coreos/ignition/internal/resource"
	"github.com/coreos/ignition/internal/util"

	"github.com/vincent-petithory/dataurl"
)

// Config represents a set of options that map to a particular OEM.
type Config struct {
	name              string
	fetch             providers.FuncFetchConfig
	newFetcher        providers.FuncNewFetcher
	baseConfig        types.Config
	defaultUserConfig types.Config
}

func (c Config) Name() string {
	return c.name
}

func (c Config) FetchFunc() providers.FuncFetchConfig {
	return c.fetch
}

func (c Config) NewFetcherFunc() providers.FuncNewFetcher {
	if c.newFetcher != nil {
		return c.newFetcher
	}
	return func(l *log.Logger, c *resource.HttpClient) (resource.Fetcher, error) {
		return resource.Fetcher{
			Logger: l,
			Client: c,
		}, nil
	}
}

func (c Config) BaseConfig() types.Config {
	return c.baseConfig
}

func (c Config) DefaultUserConfig() types.Config {
	return c.defaultUserConfig
}

var configs = registry.Create("oem configs")
var yes = util.BoolToPtr(true)

func init() {
	configs.Register(Config{
		name:  "azure",
		fetch: azure.FetchConfig,
		baseConfig: types.Config{
			Systemd: types.Systemd{
				Units: []types.Unit{
					{Enabled: yes, Name: "waagent.service"},
					{Name: "etcd2.service", Dropins: []types.Dropin{
						{Name: "10-oem.conf", Contents: "[Service]\nEnvironment=ETCD_ELECTION_TIMEOUT=1200\n"},
					}},
				},
			},
			Storage: types.Storage{Files: []types.File{serviceFromOem("waagent.service")}},
		},
		defaultUserConfig: types.Config{Systemd: types.Systemd{Units: []types.Unit{userCloudInit("Azure", "azure")}}},
	})
	configs.Register(Config{
		name:  "cloudsigma",
		fetch: noop.FetchConfig,
	})
	configs.Register(Config{
		name:  "cloudstack",
		fetch: noop.FetchConfig,
	})
	configs.Register(Config{
		name:  "digitalocean",
		fetch: digitalocean.FetchConfig,
		baseConfig: types.Config{
			Systemd: types.Systemd{
				Units: []types.Unit{{Enabled: yes, Name: "coreos-metadata-sshkeys@.service"}},
			},
		},
		defaultUserConfig: types.Config{Systemd: types.Systemd{Units: []types.Unit{userCloudInit("DigitalOcean", "digitalocean")}}},
	})
	configs.Register(Config{
		name:  "brightbox",
		fetch: openstack.FetchConfig,
		defaultUserConfig: types.Config{
			Systemd: types.Systemd{
				Units: []types.Unit{
					{Mask: true, Name: "user-configdrive.service"},
					{Mask: true, Name: "user-configvirtfs.service"},
					userCloudInit("BrightBox", "ec2-compat"),
				},
			},
		},
	})
	configs.Register(Config{
		name:  "openstack",
		fetch: openstack.FetchConfig,
		defaultUserConfig: types.Config{
			Systemd: types.Systemd{
				Units: []types.Unit{
					{Mask: true, Name: "user-configdrive.service"},
					{Mask: true, Name: "user-configvirtfs.service"},
					userCloudInit("OpenStack", "ec2-compat"),
				},
			},
		},
	})
	configs.Register(Config{
		name:       "ec2",
		fetch:      ec2.FetchConfig,
		newFetcher: ec2.NewFetcher,
		baseConfig: types.Config{
			Systemd: types.Systemd{
				Units: []types.Unit{
					{Enabled: yes, Name: "coreos-metadata-sshkeys@.service"},
					{Name: "etcd2.service", Dropins: []types.Dropin{
						{Name: "10-oem.conf", Contents: "[Service]\nEnvironment=ETCD_ELECTION_TIMEOUT=1200\n"},
					}},
				},
			},
		},
		defaultUserConfig: types.Config{
			Systemd: types.Systemd{
				Units: []types.Unit{
					{Mask: true, Name: "user-configdrive.service"},
					{Mask: true, Name: "user-configvirtfs.service"},
					userCloudInit("EC2", "ec2-compat"),
				},
			},
		},
	})
	configs.Register(Config{
		name:  "exoscale",
		fetch: noop.FetchConfig,
	})
	configs.Register(Config{
		name:  "gce",
		fetch: gce.FetchConfig,
		baseConfig: types.Config{
			Systemd: types.Systemd{
				Units: []types.Unit{
					{Enabled: yes, Name: "coreos-metadata-sshkeys@.service"},
					{Enabled: yes, Name: "oem-gce.service"},
				},
			},
			Storage: types.Storage{
				Files: []types.File{
					serviceFromOem("oem-gce.service"),
					{
						Node: types.Node{
							Filesystem: "root",
							Path:       "/etc/hosts",
						},
						FileEmbedded1: types.FileEmbedded1{
							Mode:     0444,
							Contents: contentsFromString("169.254.169.254 metadata\n127.0.0.1 localhost\n"),
						},
					},
					{
						Node: types.Node{
							Filesystem: "root",
							Path:       "/etc/profile.d/google-cloud-sdk.sh",
						},
						FileEmbedded1: types.FileEmbedded1{
							Mode: 0444,
							Contents: contentsFromString(`#!/bin/sh
alias gcloud="(docker images google/cloud-sdk || docker pull google/cloud-sdk) > /dev/null;docker run -t -i --net="host" -v $HOME/.config:/.config -v /var/run/docker.sock:/var/run/doker.sock -v /usr/bin/docker:/usr/bin/docker google/cloud-sdk gcloud"
alias gcutil="(docker images google/cloud-sdk || docker pull google/cloud-sdk) > /dev/null;docker run -t -i --net="host" -v $HOME/.config:/.config google/cloud-sdk gcutil"
alias gsutil="(docker images google/cloud-sdk || docker pull google/cloud-sdk) > /dev/null;docker run -t -i --net="host" -v $HOME/.config:/.config google/cloud-sdk gsutil"
`),
						},
					},
				},
			},
		},
		defaultUserConfig: types.Config{Systemd: types.Systemd{Units: []types.Unit{userCloudInit("GCE", "gce")}}},
	})
	configs.Register(Config{
		name:  "hyperv",
		fetch: noop.FetchConfig,
	})
	configs.Register(Config{
		name:  "niftycloud",
		fetch: noop.FetchConfig,
	})
	configs.Register(Config{
		name:  "packet",
		fetch: packet.FetchConfig,
		baseConfig: types.Config{
			Systemd: types.Systemd{
				Units: []types.Unit{
					{Enabled: yes, Name: "coreos-metadata-sshkeys@.service"},
					{Enabled: yes, Name: "packet-phone-home.service"},
				},
			},
			Storage: types.Storage{Files: []types.File{serviceFromOem("packet-phone-home.service")}},
		},
		defaultUserConfig: types.Config{Systemd: types.Systemd{Units: []types.Unit{userCloudInit("Packet", "packet")}}},
	})
	configs.Register(Config{
		name:  "pxe",
		fetch: noop.FetchConfig,
	})
	configs.Register(Config{
		name:  "rackspace",
		fetch: noop.FetchConfig,
	})
	configs.Register(Config{
		name:  "rackspace-onmetal",
		fetch: noop.FetchConfig,
	})
	configs.Register(Config{
		name:  "vagrant",
		fetch: noop.FetchConfig,
	})
	configs.Register(Config{
		name:  "vagrant-virtualbox",
		fetch: virtualbox.FetchConfig,
	})
	configs.Register(Config{
		name:  "virtualbox",
		fetch: virtualbox.FetchConfig,
	})
	configs.Register(Config{
		name:  "vmware",
		fetch: vmware.FetchConfig,
		baseConfig: types.Config{
			Systemd: types.Systemd{Units: []types.Unit{{Enabled: yes, Name: "vmtoolsd.service"}}},
			Storage: types.Storage{Files: []types.File{serviceFromOem("vmtoolsd.service")}},
		},
		defaultUserConfig: types.Config{Systemd: types.Systemd{Units: []types.Unit{userCloudInit("VMware", "vmware")}}},
	})
	configs.Register(Config{
		name:  "interoute",
		fetch: noop.FetchConfig,
	})
	configs.Register(Config{
		name:  "qemu",
		fetch: qemu.FetchConfig,
	})
	configs.Register(Config{
		name:  "file",
		fetch: file.FetchConfig,
	})
}

func Get(name string) (config Config, ok bool) {
	config, ok = configs.Get(name).(Config)
	return
}

func MustGet(name string) Config {
	if config, ok := Get(name); ok {
		return config
	} else {
		panic(fmt.Sprintf("invalid OEM name %q provided", name))
	}
}

func Names() (names []string) {
	return configs.Names()
}

func contentsFromString(data string) types.FileContents {
	return types.FileContents{
		Source: (&url.URL{
			Scheme: "data",
			Opaque: "," + dataurl.EscapeString(data),
		}).String(),
	}
}

func contentsFromOem(path string) types.FileContents {
	return types.FileContents{
		Source: (&url.URL{
			Scheme: "oem",
			Path:   path,
		}).String(),
	}
}

func userCloudInit(name string, oem string) types.Unit {
	contents := `[Unit]
Description=Cloudinit from %s metadata

[Service]
Type=oneshot
ExecStart=/usr/bin/coreos-cloudinit --oem=%s

[Install]
WantedBy=multi-user.target
`

	return types.Unit{
		Name:     "oem-cloudinit.service",
		Enabled:  yes,
		Contents: fmt.Sprintf(contents, name, oem),
	}
}

func serviceFromOem(unit string) types.File {
	return types.File{
		Node: types.Node{
			Filesystem: "root",
			Path:       "/etc/systemd/system/" + unit,
		},
		FileEmbedded1: types.FileEmbedded1{
			Mode:     0444,
			Contents: contentsFromOem("/units/" + unit),
		},
	}
}
