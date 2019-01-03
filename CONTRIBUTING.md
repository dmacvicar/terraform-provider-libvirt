# How to contribute to terraform-libvirt-plugin:

## Workflow/requirements for a PR:

- Undocumented feature do not exist.
Document each new feature.
- Untested features do not exist.
To ensure that what we code really works, relevant flows should be covered via acceptance tests.
So when thinking about a contribution, also think about testability. All tests can be run local without the need of CI. Have a look at the Testing section (later on this page)
`make` testacc will run all the tests for example.

- Creation and update resource. Consider to implement the `update` CRUD of terraform-libvirt  of an existing resource and also testing it in testacc. 
For example if an user rerun 2 times a terraform apply with a different parameter, this call will update the existing resource with the new parameter.
This step is not trivial and need some special care on the implementation. 
An example of updating a resource in testacc is here: https://github.com/dmacvicar/terraform-provider-libvirt/blob/master/libvirt/resource_libvirt_cloud_init_test.go#L73

## Conventions

* Use [commit.style](https://commit.style/) for git commit messages


## Some words about the design-architecture of this project:

Libvirt upstream use xml-schemas for defining the resources.

There is a common-pattern shared among the resources in this provider.

For example the `domain` resource, and others are organized like follow:

-  the `domain_def` contains libvirt xml schemas and operation
- `resource_libvirt_domain.go` (contains terraform CRUD operations and call the libvirt_xml)
   you can imagine the `resource` file as a sort of "main" file for each resource.
- `resource_libvirt_domain_test.go` ( contains testacc for resource)

# Testing

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

### Code coverage:

Run first the testacc suite.

Then you can visualize the profile in html format:

```golang
go tool cover -html=profile.cov
```

The codecoverage can give you more usefull infos about were you could write a new tests for improving our testacc.

Feel free to read more about this on : https://blog.golang.org/cover.

### Writing acceptance tests:

Take a look at Terraform's docs about [writing acceptance tests](https://github.com/hashicorp/terraform/blob/master/.github/CONTRIBUTING.md#writing-an-acceptance-test).

For resource names etc, use random names with the helper function. Take example from other testacc.



### Provider terraform useful devel info:

https://www.terraform.io/docs/plugins/provider.html

In addition to the terraform documentation, you might also have look at libvirt golang libraries we use.

https://godoc.org/github.com/libvirt/libvirt-go-xml
https://godoc.org/github.com/libvirt/libvirt-go


### Easy Issues for newbies:

We try to keep easy issues for new contributors with label : https://github.com/dmacvicar/terraform-provider-libvirt/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22.

Feel free to pick also other issues 
