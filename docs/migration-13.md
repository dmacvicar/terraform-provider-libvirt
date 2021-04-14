# Migrate provider to terraform v13

1) Create the following directory:

`mkdir -p ~/.local/share/terraform/plugins/registry.terraform.io/dmacvicar/libvirt/0.6.2/linux_amd64`

Where `0.6.2` is the version of the libvirt provider and `linux_amd64` is the architecture.


2) Move the provider-libvirt binary to that directory:

`mv terraform-provider-libvirt  ~/.local/share/terraform/plugins/registry.terraform.io/dmacvicar/libvirt/0.6.2/linux_amd64`

Note: you can use a [released binary](https://github.com/dmacvicar/terraform-provider-libvirt/releases), or
[build from source](https://github.com/dmacvicar/terraform-provider-libvirt#building-from-source).


3) Run an example:
https://github.com/dmacvicar/terraform-provider-libvirt/tree/master/examples/v0.13/tumbleweed

The major change from 0.12 to 0.13 is the Explicit Provider Source Locations.

With the local plugin package in place, the final step is to add a
[provider requirement](https://www.terraform.io/docs/configuration/provider-requirements.html) to each of
the modules in your configuration to state which provider you mean when you say "libvirt" elsewhere in the
module. Add the next code snippet to the `main.tf` file (including all imported modules using this libvirt
provider):

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


Further information:

https://www.terraform.io/upgrade-guides/0-13.html#in-house-providers
