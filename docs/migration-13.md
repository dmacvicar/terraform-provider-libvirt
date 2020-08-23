# Migrate provider to terraform v13


1) Create following directory:


`mkdir -p ~/.local/share/terraform/plugins/registry.terraform.io/dmacvicar/libvirt/0.6.2/linux_amd64`

Where the `0.6.2` is the version of the libvirt provider and the `linux_amd64` is the arch.


2) Move the provider-libvirt binary to that directory.


`mv terraform-provider-libvirt  ~/.local/share/terraform/plugins/registry.terraform.io/dmacvicar/libvirt/0.6.2/linux_amd64`


Note: you can use a released one binary https://github.com/dmacvicar/terraform-provider-libvirt/releases or builded from source

3) Run an example:
https://github.com/dmacvicar/terraform-provider-libvirt/tree/master/examples/v0.13/tubleweed


The major change from 0.12 to 0.13 is the Explicit Provider Source Locations.

In your `main.tf` this will be the needed change. 

```hcl

terraform {
 required_version = ">= 0.13"
  required_providers {
    libvirt = {
      source  = "dmacvicar/libvirt"
      version = "0.6.2"
    }
  }
}

```


Further infromations:

https://www.terraform.io/upgrade-guides/0-13.html#in-house-providers
