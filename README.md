# Terraform provider for libvirt

- [![Gitter chat](https://badges.gitter.im/terraform-provider-libvirt/Lobby.png)](https://gitter.im/terraform-provider-libvirt/Lobby) ([IRC gateway](https://irc.gitter.im/))
- Planning board: [Github Projects](https://github.com/dmacvicar/terraform-provider-libvirt/projects/1)


![alpha](https://img.shields.io/badge/stability%3F-beta-yellow.svg) [![Build Status](https://travis-ci.org/dmacvicar/terraform-provider-libvirt.svg?branch=master)](https://travis-ci.org/dmacvicar/terraform-provider-libvirt)
___
This is a terraform provider that lets you provision
servers on a [libvirt](https://libvirt.org/) host via [Terraform](https://terraform.io/).

## Table of Content
- [Introduction and Goals](#intro)
- [Downloading](#downloading)
- [Installing](#installing)
- [Quickstart](#using-the-provider)
- [Building from source](#building-from-source)
- [How to contribute](CONTRIBUTING.md)
- [Upstream project using this provider](#upstream-projects-using-terraform-libvirt)

## Website Docs
- [Libvirt Provider](website/docs/index.html.markdown)
- [CloudInit](website/docs/r/cloudinit.html.markdown)
- [CoreOS Ignition](website/docs/r/coreos_ignition.html.markdown)
- [Domains](website/docs/r/domain.html.markdown)
- [Networks](website/docs/r/network.markdown)
- [Volumes](website/docs/r/volume.html.markdown)

# Introduction & Goals

This project exists:

* To allow teams to get the benefits [Software Defined Infrastructure](https://en.wikipedia.org/wiki/Software-defined_infrastructure) Terraform provides, on top of classical and cheap virtualization infrastructure provided by Linux and [KVM](https://www.linux-kvm.org)
  This helps in very dynamic [DevOps](https://en.wikipedia.org/wiki/DevOps), Development and Testing activities.
* To allow for mixing KVM resources with other infrastructure Terraform is able to manage

What is *NOT* in scope:

* To support every advanced feature [libvirt](https://libvirt.org/) supports

  This would make the mapping from terraform complicated and not maintanable. See the [How to contribute](CONTRIBUTING.md) section to understand how to approach new features.
  
## Downloading

Builds for openSUSE, CentOS, Ubuntu, Fedora are created with openSUSE's [OBS](https://build.opensuse.org). The build definitions are available for both the [stable](https://build.opensuse.org/package/show/systemsmanagement:terraform/terraform-provider-libvirt) and [master](https://build.opensuse.org/package/show/systemsmanagement:terraform:unstable/terraform-provider-libvirt) branches.

## Using published binaries/builds

* *Stable releases*: Head to the [releases section](https://github.com/dmacvicar/terraform-provider-libvirt/releases) and download the latest stable release build for your distribution.
* *git master builds*: Head to the [download area of the OBS project](https://download.opensuse.org/repositories/systemsmanagement:/terraform:/unstable/) and download the build for your distribution.

## Using packages

Follow the instructions for your distribution:

* [Packages for stable releases](https://software.opensuse.org/download/package?project=systemsmanagement:terraform&package=terraform-provider-libvirt)
* [Packages for current git master](https://software.opensuse.org/download/package?project=systemsmanagement:terraform:unstable&package=terraform-provider-libvirt)

## Building from source

### Requirements

-	[Terraform](https://www.terraform.io/downloads.html)
-	[Go](https://golang.org/doc/install) (to build the provider plugin)
-	[libvirt](https://libvirt.org/downloads.html) 1.2.14 or newer development headers
-	`cgo` is required by the [libvirt-go](https://github.com/libvirt/libvirt-go) package. `export CGO_ENABLED="1"`


This project uses [go modules](https://github.com/golang/go/wiki/Modules) to vendor all its
dependencies.

You do not have to interact with `modules` since the vendored packages are **already included in the repo**.

Ensure you have the latest version of Go installed on your system, terraform usually
takes advantage of features available only inside of the latest stable release.

You need also need libvirt-dev(el) package installed.

### Building The Provider

Clone repository to: `$GOPATH/src/github.com/dmacvicar/terraform-provider-libvirt`

```console
mkdir -p $GOPATH/src/github.com/dmacvicar; cd $GOPATH/src/github.com/dmacvicar
git clone https://github.com/dmacvicar/terraform-provider-libvirt.git
```

Enter the provider directory and build the provider

```console
cd $GOPATH/src/github.com/dmacvicar/terraform-provider-libvirt
make install
```

If you are using Go >= 1.11, you don't need to build inside GOPATH:

```
export GO111MODULE=on
export GOFLAGS=-mod=vendor
make install
```

You will now find the binary at `$GOPATH/bin/terraform-provider-libvirt`.

#### Windows

To build it on Windows (64bit) one can use MinGW64 (http://www.msys2.org/)

Install Golang on Windows  
Clone terraform-provider-libvirt repository  
Open MinGW64 Console
```console
pacman -S mingw-w64-x86_64-libvirt
export PATH=$PATH:/c/Go/bin
pacman -S mingw-w64-x86_64-pkg-config
pacman -S mingw-w64-x86_64-glib2
pacman -S mingw-w64-x86_64-dbus-glib
pacman -S mingw-w64-x86_64-libssh
pacman -S mingw-w64-x86_64-yajl
export GO111MODULE=on
export GOFLAGS=-mod=vendor
go install
```

# Installing

*  Check that libvirt daemon 1.2.14 or newer is running on the hypervisor (`virsh version --daemon`)
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

You can target different libvirt hosts instantiating the [provider multiple times](https://www.terraform.io/docs/configuration/providers.html#multiple-provider-instances). [Example](examples/v0.12/multiple).


### Using qemu-agent

From its documentation, [qemu-agent](https://wiki.libvirt.org/page/Qemu_guest_agent):

>It is a daemon program running inside the domain which is supposed to help management applications with executing functions which need assistance of the guest OS.

Until terraform-provider-libvirt 0.4.2, qemu-agent was used by default to get network configuration. However, if qemu-agent is not running, this creates a delay until connecting to it times-out.

In current versions, we default to not to attempt connecting to it, and attempting to retrieve network interface information from the agent needs to be enabled explicitly with `qemu_agent = true`, further details [here](https://github.com/dmacvicar/terraform-provider-libvirt/blob/master/website/docs/r/domain.html.markdown). Note that you still need to make sure the agent is running in the OS, and that is unrelated to this option.

Note: when using bridge network configurations you need to enable the `qemu_agent = true`. otherwise you will not retrieve the ip adresses of domains. 

Be aware that this variables may be subject to change again in future versions.

## Upstream projects using terraform-libvirt:

* [sumaform](https://github.com/moio/sumaform)
   sumaform is a way to quickly configure, deploy, test [Uyuni](https://www.uyuni-project.org/) and [SUSE Manager](https://www.suse.com/products/suse-manager/) setups with clients and servers.

* [ha-cluster-sap](https://github.com/SUSE/ha-sap-terraform-deployments)
  Automated HA and SAP Deployments in Public/Private Clouds (including Libvirt/KVM)

* [ceph-open-terrarium](https://github.com/MalloZup/ceph-open-terrarium)
   ceph-open-terrarium is a way to quickly configure, deploy, tests CEPH cluster without or with [Deepsea](https://github.com/SUSE/DeepSea)

* [kubic](https://github.com/kubic-project)
    *   [kubic-terraform-kvm](https://github.com/kubic-project/kubic-terraform-kvm) Kubic Terraform script using KVM/libvirt

* [Community Driven Docker Examples](contrib/)
   Docker examples showing how to use the Libvirt Provider

* [Openshift 4 Installer](https://github.com/openshift/installer)
  The Openshift 4 Installer uses Terraform for cluster orchestration and relies on terraform-provider-libvirt for
  libvirt platform.

## Authors

* Duncan Mac-Vicar P. <dmacvicar@suse.de>

See also the list of [contributors](https://github.com/dmacvicar/terraform-provider-libvirt/graphs/contributors) who participated in this project.

The structure and boilerplate is inspired from the [Softlayer](https://github.com/finn-no/terraform-provider-softlayer) and [Google](https://github.com/terraform-providers/terraform-provider-google) Terraform provider sources.

## License

* Apache 2.0, See LICENSE file
