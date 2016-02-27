# Terraform provider for libvirt

[![experimental](http://badges.github.io/stability-badges/dist/experimental.svg)](http://github.com/badges/stability-badges)

This is a terraform provider that lets you provision
servers on a [libvirt](https://libvirt.org/) host via [Terraform](https://terraform.io/).

Build status: [![Build Status](https://travis-ci.org/dmacvicar/terraform-provider-libvirt.svg?branch=master)](https://travis-ci.org/dmacvicar/terraform-provider-libvirt)

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
    uri = ""
}
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

## Author

* Duncan Mac-Vicar P. <dmacvicar@suse.de>

## License

* Apache 2.0, See LICENSE file