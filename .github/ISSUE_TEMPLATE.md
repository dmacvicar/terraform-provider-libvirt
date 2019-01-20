##  System Information

### Linux distribution

``` openSUSE 42.2/ Centos7/ Ubuntu.. ```

### Terraform version

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

## Checklist

- [ ] Is your issue/contribution related with enabling some setting/option exposed by libvirt that the plugin does not yet support, or requires changing/extending the provider terraform schema?

  - [ ] Make sure you explain why this option is important to you, why it should be important to everyone. Describe your use-case with detail and provide examples where possible.
  - [ ] If it is a very special case, consider using the _XSLT_ support in the provider to tweak the definition instead of opening an issue
  - [ ] Maintainers do not have expertise in every libvirt setting, so please, describe the feature and how it is used. Link to the appropriate documentation

- [ ] Is it a bug or something that does not work as expected? Please make sure you fill the version information below:


## Description of Issue/Question

### Setup

(Please provide the full _main.tf_ file for reproducing the issue (Be sure to remove sensitive information)

### Steps to Reproduce Issue

(Include debug logs if possible and relevant).

___
## Additional information:

Do you have SELinux or Apparmor/Firewall enabled? Some special configuration?
Have you tried to reproduce the issue without them enabled?
