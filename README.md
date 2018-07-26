# Terraform provider for libvirt

- [![Gitter chat](https://badges.gitter.im/terraform-provider-libvirt/Lobby.png)](https://gitter.im/terraform-provider-libvirt/Lobby) ([IRC gateway](https://irc.gitter.im/))
- Planning board: [Github Projects](https://github.com/dmacvicar/terraform-provider-libvirt/projects/1)


![alpha](https://img.shields.io/badge/stability%3F-beta-yellow.svg) [![Build Status](https://travis-ci.org/dmacvicar/terraform-provider-libvirt.svg?branch=master)](https://travis-ci.org/dmacvicar/terraform-provider-libvirt) [![Coverage Status](https://coveralls.io/repos/github/dmacvicar/terraform-provider-libvirt/badge.svg?branch=master)](https://coveralls.io/github/dmacvicar/terraform-provider-libvirt?branch=master)

___
This is a terraform provider that lets you provision
servers on a [libvirt](https://libvirt.org/) host via [Terraform](https://terraform.io/).

## Table of Content
- [Downloading](#downloading)
- [Installing](#installing)
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


## Downloading

Builds for openSUSE, CentOS, Ubuntu, Fedora are created with openSUSE's [OBS](https://build.opensuse.org). The build definitions are available for both the [stable](https://build.opensuse.org/package/show/home:dmacvicar:terraform-provider-libvirt:stable/terraform-provider-libvirt) and [master](https://build.opensuse.org/project/show/home:dmacvicar:terraform-provider-libvirt) branches.

## Using released builds

* Head to the [releases section](https://github.com/dmacvicar/terraform-provider-libvirt/releases) and download the latest stable release build for your distribution.

## Using unstable builds

* Head to the [download area of the OBS project](https://download.opensuse.org/repositories/home:/dmacvicar:/terraform-provider-libvirt/) and download the build for your distribution.

## Using community packages

If you are using opensuse/SUSE distro, add the repo and download the package (check the repo according your distro)

**NOTE**: This may not be the latest release.

```console

DISTRO=openSUSE_Leap_42.3
zypper addrepo http://download.opensuse.org/repositories/Virtualization:containers/$DISTRO/Virtualization:containers.repo
zypper refresh
zypper install terraform-provider-libvirt
```

## Building from source

Before building, you will need the following

* libvirt 1.2.14 or newer development headers
* latest [golang](https://golang.org/dl/) version
* `cgo` is required by the [libvirt-go](https://github.com/libvirt/libvirt-go) package. `export CGO_ENABLED="1"`

This project uses [glide](https://github.com/Masterminds/glide) to vendor all its
dependencies.

You do not have to interact with `glide` since the vendored packages are **already included in the repo**.

Ensure you have the latest version of Go installed on your system, terraform usually
takes advantage of features available only inside of the latest stable release.

You need also need libvirt-dev(el) package installed.

```console
go get github.com/dmacvicar/terraform-provider-libvirt
cd $GOPATH/src/github.com/dmacvicar/terraform-provider-libvirt
go install
```

You will now find the binary at `$GOPATH/bin/terraform-provider-libvirt`.

# Installing

*  Check that libvirt daemon 1.2.14 or newer is running on the hypervisor
* `mkisofs` is required to use the [CloudInit](website/docs/r/cloudinit.html.markdown)

[Copied from the Terraform documentation](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins):

At present Terraform can automatically install only the providers distributed by HashiCorp. Third-party providers can be manually installed by placing their plugin executables in one of the following locations depending on the host operating system:

> On Linux and unix systems, in the sub-path `.terraform.d/plugins` in your user's home directory.

> On Windows, in the sub-path `terraform.d/plugins` beneath your user's "Application Data" directory.

terraform init will search this directory for additional plugins during plugin initialization.

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

### Using multiple hypervisors / provider instances

You can target different libvirt hosts instantiating the [provider multiple times](https://www.terraform.io/docs/configuration/providers.html#multiple-provider-instances). [Example](examples/multiple).

## Troubleshooting (aka you have a problem)

Have a look at [TROUBLESHOOTING](doc/TROUBLESHOOTING.md), and feel free to add a PR if you find out something is missing.

## Authors

* Duncan Mac-Vicar P. <dmacvicar@suse.de>

See also the list of [contributors](https://github.com/dmacvicar/terraform-provider-libvirt/graphs/contributors) who participated in this project.

The structure and boilerplate is inspired from the [Softlayer](https://github.com/finn-no/terraform-provider-softlayer) and [Google](https://github.com/terraform-providers/terraform-provider-google) Terraform provider sources.

## License

* Apache 2.0, See LICENSE file
