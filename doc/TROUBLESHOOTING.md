# Troubleshooting

## You cannot build from source

You run 

```console
glide install
go build 
```
And you get:

```console
Package libvirt was not found in the pkg-config search path.
Perhaps you should add the directory containing `libvirt.pc'
to the PKG_CONFIG_PATH environment variable
No package 'libvirt' found
pkg-config: exit status 1
```

You probably need libvirt-dev(el) package installed.

```console 
apt install libvirt-dev
```

## Bug in libvirt

* There is a [bug in libvirt](https://bugzilla.redhat.com/show_bug.cgi?id=1293804) that seems to be causing
  problems to unlink volumes. Tracked [here](https://github.com/dmacvicar/terraform-provider-libvirt/issues/6).

  If you see something like:

  ```console
  cannot unlink file '/var/lib/libvirt/images/XXXXXXXXXXXX': Permission denied
  ```
  It is probably related and fixed in libvirt 1.3.3 (already available in openSUSE Tumbleweed).

## Selinux on Debian distros

* On Debian distros SELinux is enforced by qemu even if it is disabled globally, this might cause unexpected `Could not open '/var/lib/libvirt/images/<FILE_NAME>': Permission denied` errors. Double check that `security_driver = "none"` is uncommented in `/etc/libvirt/qemu.conf` and issue `sudo systemctl restart libvirt-bin` to restart the daemon.


## Cannot get terraform (building from src)

When trying to build the pkg from source, you get **cannot find package "context"
´´´console
go build
../../go/src/github.com/dmacvicar/terraform-provider-libvirt/vendor/github.com/hashicorp/terraform/terraform/context.go:4:2: cannot find package "context" in any of:
    ...
´´´
Install the latest golang version. 

https://github.com/hashicorp/terraform/issues/12470
