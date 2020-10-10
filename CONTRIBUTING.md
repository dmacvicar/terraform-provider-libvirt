# How to contribute to terraform-provider-libvirt

## Checklist

### New features

  - Is your issue/contribution related with enabling some setting/option exposed by libvirt that the plugin does not yet support, or requires changing/extending the provider terraform schema?
    - [ ] Start with discussion/issue to discuss the enthusiasm of the team about the enhancement and discuss different approaches and mention @dmacvicar.
    - [ ] Make sure you explain why this option is important to you, why it should be important to everyone. Describe your use-case with detail and provide examples where possible.
    - [ ] If it is a very special case, consider using the _XSLT_ support in the provider to tweak the definition instead of opening an issue
  - [ ] Does the feature you added include documentation in the [expected place](https://github.com/dmacvicar/terraform-provider-libvirt/tree/master/website/docs)?
  - [ ] Does your feature include the appropriate tests?

### Bugfixes

- Does this fix a bug or something that does not work as expected?
  - [ ] If possible, start the fix with a single testcase reproducing the issue in a separate commit
  - [ ] If there is an issue open, please mention it in the Pull Request
  - [ ] If there is not an issue open, consider including information on where it happens and how to reproduce it, following the [issues template](.github/ISSUE_TEMPLATE.md)
- [ ] Maintainers do not have expertise in every libvirt setting, so please, describe how the new or current feature works and how it is used. Link to the appropriate documentation
- [ ] Does your PR follow the conventions below?

## Implementation notes

- Creation and update resource. Consider to implement the `update` CRUD of terraform-libvirt  of an existing resource and also testing it in acceptance tests.
For example if an user rerun 2 times a terraform apply with a different parameter, this call will update the existing resource with the new parameter.
This step is not trivial and need some special care on the implementation. 
An example of updating a resource in acceptance tests is here: https://github.com/dmacvicar/terraform-provider-libvirt/blob/master/libvirt/resource_libvirt_cloud_init_test.go#L73

## Conventions

* Use [commit.style](https://commit.style/) for git commit messages

## Some words about the design-architecture of this project:

Libvirt upstream use xml-schemas for defining the resources.

There is a common-pattern shared among the resources in this provider.

For example the `domain` resource, and others are organized like follow:

-  the `domain_def` contains libvirt xml schemas and operation
- `resource_libvirt_domain.go` (contains terraform CRUD operations and call the libvirt_xml)
   you can imagine the `resource` file as a sort of "main" file for each resource.
- `resource_libvirt_domain_test.go` ( contains acceptance tests for resource)

## Testing

To ensure that what we code really works, relevant flows should be covered via acceptance tests.
So when thinking about a contribution, also think about testability. All tests can be run local without the need of CI. Have a look at the Testing section (later on this page)
`make` testacc will run all the tests for example.

## Running the tests (testacc)

```
make testacc
```

You can also run some particular test with:

```
make testacc TEST_ARGS="-run TestAccLibvirtDomain_Cpu"
```

Or run a group of test with a verbose loglevel:

```bash
TF_LOG=DEBUG make testacc TEST_ARGS="-run TestAccLibvirtNet*"
```

If you run the tests on an unprivileged connection (e.g session libvirt), some of the acceptance tests will need to be disabled (mainly the networking ones) through an environment
variable:

```bash
LIBVIRT_DEFAULT_URI='qemu+unix:///session' TF_LIBVIRT_DISABLE_PRIVILEGED_TESTS=1 make testacc
```

If '/dev/random' is not available on the platform you run the acceptance tests on, you can override the device used
through an environment variable as well:

```bash
TF_LIBVIRT_RNG_DEV='/dev/random' make testacc
```

### Code coverage:

Run first the testacc suite.

Then you can visualize the profile in html format:

```golang
go tool cover -html=profile.cov
```

The codecoverage can give you more usefull infos about were you could write a new tests for improving our acceptance tests.

Feel free to read more about this on : https://blog.golang.org/cover.

### Writing acceptance tests:

Take a look at Terraform's docs about [writing acceptance tests](https://github.com/hashicorp/terraform/blob/master/.github/CONTRIBUTING.md#writing-an-acceptance-test).

For resource names etc, use random names with the helper function. Take example from other acceptance tests.

## Other learning resources

https://www.terraform.io/docs/plugins/provider.html

In addition to the terraform documentation, you might also have look at libvirt golang libraries we use.

https://godoc.org/github.com/libvirt/libvirt-go-xml
https://godoc.org/github.com/libvirt/libvirt-go


## Easy issues for newbies:

We try to keep easy issues for new contributors with label : https://github.com/dmacvicar/terraform-provider-libvirt/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22.

Feel free to pick also other issues 
