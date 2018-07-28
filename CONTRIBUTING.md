# How to contribute to terraform-libvirt-plugin:

## Running the tests (testacc)

```
make testacc
```

### code coverage:

Run first the testacc suite.

Then you can visualize the profile in html format:

```golang
go tool cover -html=profile.cov
```

Feel free to read more about this on : https://blog.golang.org/cover

### Provider terraform useful devel info:

https://www.terraform.io/docs/plugins/provider.html

### Writing acceptance tests:

Take a look at Terraform's docs about [writing acceptance tests](https://github.com/hashicorp/terraform/blob/master/.github/CONTRIBUTING.md#writing-an-acceptance-test).
