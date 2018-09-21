
This example takes a base image of 2G and uses it as the backing store of a bigger volume. _cloud-init_ is then used to resize the original partition to the new available space.

See https://github.com/dmacvicar/terraform-provider-libvirt/issues/369

# Running

```console
$ terraform apply
$ ssh -i id_rsa "root@$(terraform output ip)"
```
