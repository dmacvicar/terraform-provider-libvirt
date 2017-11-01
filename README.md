# Terraform provider for libvirt

![alpha](https://img.shields.io/badge/stability%3F-beta-yellow.svg) [![Build Status](https://travis-ci.org/dmacvicar/terraform-provider-libvirt.svg?branch=master)](https://travis-ci.org/dmacvicar/terraform-provider-libvirt) [![Coverage Status](https://coveralls.io/repos/github/dmacvicar/terraform-provider-libvirt/badge.svg?branch=master)](https://coveralls.io/github/dmacvicar/terraform-provider-libvirt?branch=master)

This provider is still being actively developed. To see what is left or planned, see the [issues list](https://github.com/dmacvicar/terraform-provider-libvirt/issues).

This is a terraform provider that lets you provision
servers on a [libvirt](https://libvirt.org/) host via [Terraform](https://terraform.io/).

## Requirements

* libvirt 1.2.14 or newer on the hypervisor

The provider uses `virDomainInterfaceAddresses` which was added in 1.2.14. If you need a stable server distribution with a recent libvirt version, try [openSUSE Leap](https://www.opensuse.org/) or [Ubuntu](http://www.ubuntu.com/server) (from version 15.10 Wily Werewolf on).

In the future, I may try to support older libvirt versions if I find a way to elegantely conditional compile the code and get the IP addresses with alternative methods.

## Installing

[Copied from the Terraform documentation](https://www.terraform.io/docs/plugins/basics.html):
> To install a plugin, put the binary somewhere on your filesystem, then configure Terraform to be able to find it. The configuration where plugins are defined is ~/.terraformrc for Unix-like systems and %APPDATA%/terraform.rc for Windows.

If you are using opensuse/SUSE distro, add the repo and download the package (check the repo according your distro)

```console

DISTRO=openSUSE_Leap_42.1
zypper addrepo http://download.opensuse.org/repositories/Virtualization:containers/$DISTRO/Virtualization:containers.repo
zypper refresh
zypper install terraform-provider-libvirt

```

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
$ terraform plan
$ terraform apply
$ terraform destroy
```

## Building from source

This project uses [glide](https://github.com/Masterminds/glide) to vendor all its
dependencies.

Ensure you have latest version of Go installed on your system, terraform usually
takes advantage of fetures available only inside of the latest stable release.

You need also need libvirt-dev(el) package installed.

Run `go install` to build the binary. You will now find the binary at
`$GOPATH/bin/terraform-provider-libvirt`.

## Running

1.  create the example file libvirt.tf in your working directory
2.  terraform plan
3.  terraform apply

## Running acceptance tests

You need to define the LIBVIRT_DEFAULT_URI and TF_ACC variables:

```console
export LIBVIRT_DEFAULT_URI=qemu:///system
export TF_ACC=1
go test ./...
```

## Troubleshooting (aka you have a problem)

Have a look at [TROUBLESHOOTING](doc/TROUBLESHOOTING.md), and feel free to add a PR if you find out something is missing.

## Authors

* Duncan Mac-Vicar P. <dmacvicar@suse.de>

See also the list of [contributors](https://github.com/dmacvicar/terraform-provider-libvirt/graphs/contributors) who participated in this project.

The structure and boilerplate is inspired from the [Softlayer](https://github.com/finn-no/terraform-provider-softlayer) and [Google](https://github.com/hashicorp/terraform/tree/master/builtin/providers/google) Terraform provider sources.

## License

* Apache 2.0, See LICENSE file
