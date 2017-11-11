# Where to start?

Look at open issues, especially with label:

- [junior job](https://github.com/dmacvicar/terraform-provider-libvirt/issues?q=is%3Aissue+is%3Aopen+label%3A%22junior+job%22), 
- [help wanted](https://github.com/dmacvicar/terraform-provider-libvirt/issues?q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22)

### Makefile devel workflow.

Go in your build dir ```cd $GOPATH/src/github.com/dmacvicar/terraform-provider-libvirt```,
then  use

```console
make test
```

# Contribute

1. If you have issues, check out the [troubleshooting](https://github.com/dmacvicar/terraform-provider-libvirt/blob/master/doc/TROUBLESHOOTING.md)
2. Do your code (reference issue on your pr if you fix them) [Closing issues keywords](https://help.github.com/articles/closing-issues-using-keywords/)
3. Use the Makefile workflow before submit the PR (building, tests).
4. Test your code by running the acceptance tests ```make test```
5. Create a PR

## Writing acceptance tests.

Look at 
https://github.com/hashicorp/terraform/blob/master/.github/CONTRIBUTING.md#running-an-acceptance-test

## Running acceptance tests 

```console
make testacc
```
