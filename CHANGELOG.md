## 0.5.1 (December 14, 2018)

### HIGHLIGHTS:

- The provider will retry HTTP downloads if they were caused by server errors (https://github.com/dmacvicar/terraform-provider-libvirt/pull/479)
- XSLT can be used to transform the libvirt XML definition before creating resources, allowing to enable features the provider does not support (https://github.com/dmacvicar/terraform-provider-libvirt/pull/431)
- volumes: when the _HEAD_ method is forbidden, the provider will try a body-less _GET_ instead (https://github.com/dmacvicar/terraform-provider-libvirt/pull/472)
- network: add support for updating _dns.hosts_ (https://github.com/dmacvicar/terraform-provider-libvirt/pull/469)
- network: add support for setting _SRV_ entries (https://github.com/dmacvicar/terraform-provider-libvirt/pull/460)
- qemu-agent: do not contact the qemu agent if the domain is shutdown (https://github.com/dmacvicar/terraform-provider-libvirt/pull/474)
- cli:  add `-version` flag (https://github.com/dmacvicar/terraform-provider-libvirt/pull/444)  

## 0.5 (October 10, 2018)

### HIGHLIGHTS:

#### libvirt Domain

* _cloud_init_ resource is obsolete and replaced with _cloud_init_disk_ (https://github.com/dmacvicar/terraform-provider-libvirt/pull/410)

  The resource does not allow individual fields anymore, and directly takes the _user_data_, _network_config_ and _meta_data_ fields, which you can provide directly or reading from a file, optionally using a _template_ resource and the rendered function. See the [updated documentation](https://github.com/dmacvicar/terraform-provider-libvirt/blob/master/website/docs/r/cloudinit.html.markdown).

#### Volumes/Disk/Storage

* Allow to set the size of a volume larger than its backing volume, which allows to resize the partition then using _cloud-init_, and not be limited by the backing image (#369 and #357).

## 0.4.4 (September 16, 2018)

### HIGHLIGHTS:

#### libvirt Domain

* `TF_USE_QEMU_AGENT` variable is deprecated and replaced by a domain property `qemu_agent`.
  Because variables can be interpolated into properties, and variables can be [passed  via environment variables](https://www.terraform.io/docs/configuration/environment-variables.html#tf_var_name), the old behavior can be emulated.

#### Volumes/Disk/Storage

* Automatic disk driver selection based on volume format and automatic volume format detection(https://github.com/dmacvicar/terraform-provider-libvirt/commit/676b5a3fec75664990e5a91f24859f35becdee6a)

#### Networking

* `dhcp` paramater is an optional parameter now, disabled by default. (https://github.com/dmacvicar/terraform-provider-libvirt/pull/385)
* DNS forwarders were reworked. `localonly` option was added to libvirt-network (https://github.com/dmacvicar/terraform-provider-libvirt/commit/7651ee5824f77f0c7485736315d5a24762f85e60)
* A datasource called `libvirt_network_dns_hosts_template` can be used to populate the `dns_host` attribute in `libvirt_network` resources. (https://github.com/dmacvicar/terraform-provider-libvirt/commit/a4d0ba6a319d8728cb5d6c10aae593bdd27da516)
___
#### General improvements

* Acceptance tests are now idempotent (no dependency between resource of various tests), which avoids cascade failures. (several PRs and commits)
* Project dependencies were updated ( https://github.com/dmacvicar/terraform-provider-libvirt/commit/1347e7cabbe68d93f7cc065339636854d7c7d340)
* The error message when uploading a volume fails was improved (https://github.com/dmacvicar/terraform-provider-libvirt/commit/1aec44e0c990c4edb22578125bae33f92c4a4f39)
___
#### Bugs

* `netIface["bridge"]` now uses the correct value (https://github.com/dmacvicar/terraform-provider-libvirt/commit/2e93c78b2aea17b48639b3d613f12bfad851fd52) 

## 0.4.3 (August 14, 2018)

HIGHLIGHTS:

* *IMPORTANT* qemu-agent is not used by default to gather network
  interface information anymore. If you need to use, please set
  the TF_USE_QEMU_AGENT environment variable.

* Handle gracefully out-of-band destruction of volume and cloud-init
  resources. Should provide a better end-user experience in day to day
  operations.

## 0.4.2 (August 3, 2018)

HIGHLIGHTS:

* Fix crashes when using network devices not associated with a
  network name (regression introduced in 0.4)

## 0.4.1 (July 28, 2018)

HIGHLIGHTS:

* Fix broken ip address detection bug that was introduced in 0.4
* Add support for importing domain, network, volumes (#336)

## 0.4 (July 25, 2018)

HIGHLIGHTS:

* Support for multiple provider instances (ie: hypervisors) with different URIs
* Support for keyword-less and nested equal signs in kernel parameters
* Adds the `running` attribute when creating a domain
* Fix a bug with UEFI/OVMF booting on remote hypervisors
* Update the project dependencies to more recent versions
* The project now provides builds
* The project now has a gitter.im channel
* Integration tests are fixed and working again


