# How to contribute to terraform-libvirt-plugin:

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
