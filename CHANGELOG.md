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


