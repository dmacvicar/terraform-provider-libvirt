#!/bin/bash
set -x

unset http_proxy
export TERRAFORM_LIBVIRT_TEST_DOMAIN_TYPE=qemu
export LIBVIRT_DEFAULT_URI="qemu:///system"
export TF_ACC=true
go test -v -covermode=count -coverprofile=profile.cov -timeout=1200s ./libvirt
