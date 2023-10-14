# Terraform provider for libvirt

- [![Gitter chat](https://badges.gitter.im/terraform-provider-libvirt/Lobby.png)](https://gitter.im/terraform-provider-libvirt/Lobby) ([IRC gateway](https://irc.gitter.im/))
- Planning board: [Github Projects](https://github.com/dmacvicar/terraform-provider-libvirt/projects/1)


![alpha](https://img.shields.io/badge/stability%3F-beta-yellow.svg) [![Tests](https://github.com/dmacvicar/terraform-provider-libvirt/actions/workflows/test.yml/badge.svg)](https://github.com/dmacvicar/terraform-provider-libvirt/actions/workflows/test.yml) [![Registry](https://img.shields.io/badge/libvirt-Terraform%20Registry-blue)](https://registry.terraform.io/providers/dmacvicar/libvirt/latest/docs)

___
This is a terraform provider that lets you provision
servers on a [libvirt](https://libvirt.org/) host via [Terraform](https://terraform.io/).

## Introduction & Goals

This project exists:

* To allow teams to get the benefits [Software Defined Infrastructure](https://en.wikipedia.org/wiki/Software-defined_infrastructure) Terraform provides, on top of classical and cheap virtualization infrastructure provided by Linux and [KVM](https://www.linux-kvm.org)
  This helps in very dynamic [DevOps](https://en.wikipedia.org/wiki/DevOps), Development and Testing activities.
* To allow for mixing KVM resources with other infrastructure Terraform is able to manage

What is *NOT* in scope:

* To support every advanced feature [libvirt](https://libvirt.org/) supports

  This would make the mapping from terraform complicated and not maintainable. See the [How to contribute](CONTRIBUTING.md) section to understand how to approach new features.

## Getting started

The provider is available for auto-installation from the [Terraform Registry](https://registry.terraform.io/providers/dmacvicar/libvirt/latest).

In your `main.tf` file, specify the version you want to use:

```hcl
terraform {
  required_providers {
    libvirt = {
      source = "dmacvicar/libvirt"
    }
  }
}

provider "libvirt" {
  # Configuration options
}
```

And now run terraform init:

```console
$ terraform init
```

### Creating your first virtual machine

Here is an example that will setup the following:

+ A virtual server resource

(create this as main.tf and run terraform commands from this directory):
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

Look at more advanced examples [here](examples/) and check the [documentation](https://registry.terraform.io/providers/dmacvicar/libvirt/latest/docs).

## Manual installation

You can also manually download the provider from the [releases section](https://github.com/dmacvicar/terraform-provider-libvirt/releases) on Github. To install it, refer to the [Terraform documentation](https://www.terraform.io/docs/cli/config/config-file.html#provider-installation).

## Building from source

-	[Go](https://golang.org/doc/install) is required for building.

```bash
git clone https://github.com/dmacvicar/terraform-provider-libvirt.git
cd terraform-provider-libvirt
make
```

The binary will be called `terraform-provider-libvirt`.

### Using multiple hypervisors / provider instances

You can target different libvirt hosts instantiating the [provider multiple times](https://www.terraform.io/docs/configuration/providers.html#multiple-provider-instances). [Example](examples/v0.12/multiple).

### Using qemu-agent

From its documentation, [qemu-agent](https://wiki.libvirt.org/page/Qemu_guest_agent):

>It is a daemon program running inside the domain which is supposed to help management applications with executing functions which need assistance of the guest OS.

Until terraform-provider-libvirt 0.4.2, qemu-agent was used by default to get network configuration. However, if qemu-agent is not running, this creates a delay until connecting to it times-out.

In current versions, we default to not to attempt connecting to it, and attempting to retrieve network interface information from the agent needs to be enabled explicitly with `qemu_agent = true`, further details [here](https://github.com/dmacvicar/terraform-provider-libvirt/blob/master/website/docs/r/domain.html.markdown). Note that you still need to make sure the agent is running in the OS, and that is unrelated to this option.

Note: when using bridge network configurations you need to enable the `qemu_agent = true`. otherwise you will not retrieve the ip addresses of domains. 

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
  
* [Kubitect](https://github.com/MusicDin/kubitect) - a CLI tool for deploying and managing Kubernetes clusters on libvirt platform.

## Authors

* Duncan Mac-Vicar P. <duncan@mac-vicar.eu>

See also the list of [contributors](https://github.com/dmacvicar/terraform-provider-libvirt/graphs/contributors) who participated in this project.

The structure and boilerplate is inspired from the [Softlayer](https://github.com/finn-no/terraform-provider-softlayer) and [Google](https://github.com/terraform-providers/terraform-provider-google) Terraform provider sources.

## License

* Apache 2.0, See LICENSE file
