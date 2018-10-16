#  Version Reports:

### Distro  version of host:

``` openSUSE 42.2/ Centos7/ Ubuntu.. ```

### Terraform Version Report

```sh
terraform -v
```

### Provider and libvirt versions

```sh
terraform-provider-libvirt -version
```

If that gives you "was not built correctly", get the Git commit hash from your local provider repository:

```sh
git describe --always --abbrev=40 --dirty
```
___
# Description of Issue/Question

### Setup
(Please provide the full **main.tf** file for reproducing the issue (Be sure to remove sensitive info)

### Steps to Reproduce Issue
(Include debug logs if possible and relevant.)

___
# Additional Infos:

Do you have SELinux or Apparmor/Firewall enabled? Some special configuration?
Have you tried to reproduce the issue without them enabled?
