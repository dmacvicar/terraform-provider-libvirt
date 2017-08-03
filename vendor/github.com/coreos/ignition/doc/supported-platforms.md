# Supported Platforms #

Ignition is currently only supported for the following platforms:

* [Bare Metal] - Use the `coreos.config.url` kernel parameter to provide a URL to the configuration. The URL can use the `http://` or `tftp://` schemes to specify a remote config or the `oem://` scheme to specify a local config, rooted in `/usr/share/oem`.
* [PXE] - Use the `coreos.config.url` and `coreos.first_boot=1` (**in case of the very first PXE boot only**) kernel parameters to provide a URL to the configuration. The URL can use the `http://` or `tftp://` schemes to specify a remote config or the `oem://` scheme to specify a local config, rooted in `/usr/share/oem`.
* [Amazon EC2] - Ignition will read its configuration from the instance userdata. SSH keys are handled by coreos-metadata.
* [Microsoft Azure] - Ignition will read its configuration from the custom data provided to the instance. SSH keys are handled by the Azure Linux Agent.
* [VMware] - Use the VMware Guestinfo variables `coreos.config.data` and `coreos.config.data.encoding` to provide the config and its encoding to the virtual machine. Valid encodings are "", "base64", and "gzip+base64". Guestinfo variables can be provided directly or via an OVF environment.
* [Google Compute Engine] - Ignition will read its configuration from the instance metadata entry named "user-data". SSH keys are handled by coreos-metadata.
* [Packet] - Ignition will read its configuration from the instance userdata. SSH keys are handled by coreos-metadata.
* [QEMU] - Ignition will read its configuration from the 'opt/com.coreos/config' key on the QEMU Firmware Configuration Device.
* [DigitalOcean] - Ignition will read its configuration from the droplet userdata. SSH keys and network configuration are handled by coreos-metadata.

Ignition is under active development so expect this list to expand in the coming months.

[Bare Metal]: https://github.com/coreos/docs/blob/master/os/installing-to-disk.md
[PXE]: https://github.com/coreos/docs/blob/master/os/booting-with-pxe.md
[Amazon EC2]: https://github.com/coreos/docs/blob/master/os/booting-on-ec2.md
[Microsoft Azure]: https://github.com/coreos/docs/blob/master/os/booting-on-azure.md
[VMware]: https://github.com/coreos/docs/blob/master/os/booting-on-vmware.md
[Google Compute Engine]: https://github.com/coreos/docs/blob/master/os/booting-on-google-compute-engine.md
[Packet]: https://github.com/coreos/docs/blob/master/os/booting-on-packet.md
[QEMU]: https://github.com/qemu/qemu/blob/d75aa4372f0414c9960534026a562b0302fcff29/docs/specs/fw_cfg.txt
[DigitalOcean]: https://github.com/coreos/docs/blob/master/os/booting-on-digitalocean.md
