# Terraform provider for libvirt

![experimental](https://img.shields.io/badge/stability%3F-experimental-orange.svg) [![Build Status](https://travis-ci.org/dmacvicar/terraform-provider-libvirt.svg?branch=master)](https://travis-ci.org/dmacvicar/terraform-provider-libvirt)

This provider is still experimental/in development. To see what is left or planned, see the [issues list](https://github.com/dmacvicar/terraform-provider-libvirt/issues).

This is a terraform provider that lets you provision
servers on a [libvirt](https://libvirt.org/) host via [Terraform](https://terraform.io/).

## Requirements

* libvirt 1.2.14 on the hypervisor

The provider uses `virDomainInterfaceAddresses` which was added in 1.2.14. Distributions like Ubuntu 14.04 LTS ship 1.2.2. If you need a stable server distribution with a recent libvirt version, try [openSUSE Leap](https://www.opensuse.org/).

In the future, I may try to support older libvirt versions if I find a way to elegantely conditional compile the code and get the IP addresses with alternative methods.

## Installing

[Copied from the Terraform documentation](https://www.terraform.io/docs/plugins/basics.html):
> To install a plugin, put the binary somewhere on your filesystem, then configure Terraform to be able to find it. The configuration where plugins are defined is ~/.terraformrc for Unix-like systems and %APPDATA%/terraform.rc for Windows.

The binary should be renamed to terraform-provider-libvirt

You should update your .terraformrc and refer to the binary:

```hcl
providers {
  libvirt = "/path/to/terraform-provider-libvirt"
}
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

1.  [Install Go](https://golang.org/doc/install) on your machine
2.  [Set up Gopath](https://golang.org/doc/code.html)
3.  `git clone` this repository into `$GOPATH/src/github.com/dmacvicar/terraform-provider-libvirt`
4.  Run `go get` to get dependencies
5.  Run `go install` to build the binary. You will now find the
    binary at `$GOPATH/bin/terraform-provider-libvirt`.

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

## Known Problems

* There is a [bug in libvirt](https://bugzilla.redhat.com/show_bug.cgi?id=1293804) that seems to be causing
  problems to unlink volumes. Tracked [here](https://github.com/dmacvicar/terraform-provider-libvirt/issues/6).

  If you see something like:

  ```console
  cannot unlink file '/var/lib/libvirt/images/XXXXXXXXXXXX': Permission denied
  ```
  It is probably related and fixed in libvirt 1.3.3 (already available in openSUSE Tumbleweed).

## Author

* Duncan Mac-Vicar P. <dmacvicar@suse.de>

The structure and boilerplate is inspired from the [Softlayer](https://github.com/finn-no/terraform-provider-softlayer) and [Google](https://github.com/hashicorp/terraform/tree/master/builtin/providers/google) Terraform provider sources.

## License

* Apache 2.0, See LICENSE file
