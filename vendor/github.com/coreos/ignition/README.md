# Ignition

Ignition is the utility used by CoreOS Container Linux to manipulate disks during the initramfs. This includes partitioning disks, formatting partitions, writing files (regular files, systemd units, networkd units, etc.), and configuring users. On first boot, Ignition reads its configuration from a source of truth (remote URL, network metadata service, hypervisor bridge, etc.) and applies the configuration.

## Usage

Odds are good that you don't want to invoke Ignition directly. In fact, it isn't even present in the Container Linux root filesystem. Take a look at the [Getting Started Guide][getting started] for details on providing Ignition with a runtime configuration.

Use the [bug tracker][issues] to report bugs.

## Config Validation

To validate a config for Ignition there are binaries for a cli tool called ignition-validate available [on the releases page][releases], and an online validator available [on the CoreOS website][online-validator].

[getting started]: doc/getting-started.md
[issues]: https://github.com/coreos/bugs/issues/new?labels=component/ignition
[releases]: https://github.com/coreos/ignition/releases
[online-validator]: https://coreos.com/validate/
