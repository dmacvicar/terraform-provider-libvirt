# Terraform provider for libvirt

![alpha](https://img.shields.io/badge/stability%3F-beta-yellow.svg) [![Build Status](https://travis-ci.org/dmacvicar/terraform-provider-libvirt.svg?branch=master)](https://travis-ci.org/dmacvicar/terraform-provider-libvirt) [![Coverage Status](https://coveralls.io/repos/github/dmacvicar/terraform-provider-libvirt/badge.svg?branch=master)](https://coveralls.io/github/dmacvicar/terraform-provider-libvirt?branch=master)

___
This is a terraform provider that lets you provision
servers on a [libvirt](https://libvirt.org/) host via [Terraform](https://terraform.io/).

## Table of Content
- [Installing](#Installing)
- [Quickstart](#using-the-provider)
- [Building from source](#building-from-source)
- [How to contribute](doc/CONTRIBUTING.md)

## Website Docs
- [Libvirt Provider](website/docs/index.html.markdown)
- [CloudInit](website/docs/r/cloudinit.html.markdown)
- [CoreOS Ignition](website/docs/r/coreos_ignition.html.markdown)
- [Domains](website/docs/r/domain.html.markdown)
- [Networks](website/docs/r/network.markdown)
- [Volumes](website/docs/r/volume.html.markdown)


## Installing

###### Requirements

* libvirt 1.2.14 or newer on the hypervisor
* latest [golang](https://golang.org/dl/) version

[Copied from the Terraform documentation](https://www.terraform.io/docs/plugins/basics.html):
> To install a plugin, put the binary somewhere on your filesystem, then configure Terraform to be able to find it. The configuration where plugins are defined is ~/.terraformrc for Unix-like systems and %APPDATA%/terraform.rc for Windows.

If you are using opensuse/SUSE distro, add the repo and download the package (check the repo according your distro)
```console

DISTRO=openSUSE_Leap_42.3
zypper addrepo http://download.opensuse.org/repositories/Virtualization:containers/$DISTRO/Virtualization:containers.repo
zypper refresh
zypper install terraform-provider-libvirt
```

On debian systems you can use ```alien``` for converting the rpms

## Building from source

This project uses [glide](https://github.com/Masterminds/glide) to vendor all its
dependencies.

You do not have to interact with `glide` since the vendored packages are **already included in the repo**.

Ensure you have the latest version of Go installed on your system, terraform usually
takes advantage of features available only inside of the latest stable release.

You need also need libvirt-dev(el) package installed.

Run `go install` to build the binary. You will now find the binary at
`$GOPATH/bin/terraform-provider-libvirt`.


## Using the provider

Here is an example that will setup the following:

+ A virtual server resource

(create this as libvirt.tf and run terraform commands from this directory):
```hcl
provider "libvirt" {
    uri = "qemu:///system"
}
```

You can also set the URI in the LIBVIRT_DEFAULT_URI environment variable.

Now, define a libvirt domain:

```hcl
resource "libvirt_domain" "terraform_test" {
  name = "terraform_test"
}
```

Now you can see the plan, apply it, and then destroy the infrastructure:

```console
$ terraform init
$ terraform plan
$ terraform apply
$ terraform destroy
```

Look at more advanced examples [here](examples/)

## Troubleshooting (aka you have a problem)

Have a look at [TROUBLESHOOTING](doc/TROUBLESHOOTING.md), and feel free to add a PR if you find out something is missing.

## Authors

* Duncan Mac-Vicar P. <dmacvicar@suse.de>

See also the list of [contributors](https://github.com/dmacvicar/terraform-provider-libvirt/graphs/contributors) who participated in this project.

The structure and boilerplate is inspired from the [Softlayer](https://github.com/finn-no/terraform-provider-softlayer) and [Google](https://github.com/terraform-providers/terraform-provider-google) Terraform provider sources.

## License

* Apache 2.0, See LICENSE file
