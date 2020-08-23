# Migrate provider to terraform v13


1) Create following directory:


`mkdir -p ~/.local/share/terraform/plugins/registry.terraform.io/dmacvicar/libvirt/0.6.2/linux_amd64`

Where the 0.6.2 is the version of the libvirt provider and the last dir is the arch.

2) Move the provider binary to that directory (either a released one https://github.com/dmacvicar/terraform-provider-libvirt/releases or builded from source)

`mv terraform-provider-libvirt  ~/.local/share/terraform/plugins/registry.terraform.io/dmacvicar/libvirt/0.6.2/linux_amd64`

3) Run an example:




https://www.terraform.io/upgrade-guides/0-13.html#in-house-providers
