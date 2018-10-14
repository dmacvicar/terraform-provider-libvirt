# How to contribute to terraform-libvirt-plugin:

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


## Running the tests (testacc)

```
make testacc
```

### Code coverage:

Run first the testacc suite.

Then you can visualize the profile in html format:

```golang
go tool cover -html=profile.cov
```

The codecoverage can give you more usefull infos about were you could write a new tests for improving our testacc.

Feel free to read more about this on : https://blog.golang.org/cover.


### Provider terraform useful devel info:

https://www.terraform.io/docs/plugins/provider.html

In addition to the terraform documentation, you might also have look at libvirt golang libraries we use.

https://godoc.org/github.com/libvirt/libvirt-go-xml
https://godoc.org/github.com/libvirt/libvirt-go


### Writing acceptance tests:

Take a look at Terraform's docs about [writing acceptance tests](https://github.com/hashicorp/terraform/blob/master/.github/CONTRIBUTING.md#writing-an-acceptance-test).

For resource names etc, use random names with the helper function. Take example from other testacc.


### Easy Issues for newbies:

We try to keep easy issues for new contributors with label : https://github.com/dmacvicar/terraform-provider-libvirt/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22.

Feel free to pick also other issues 
