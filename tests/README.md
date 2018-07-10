## easy integration tests for terraform-libvirt-plugin

Which problem will adress:

the current unit-tests/acceptance are not longterm maintenable and even are difficult to write for newcomer.
Would be really easy if any user could provide a terraform.tf file and add it to a yml file, without having to worry to much about terraform internals.

## todos:

0) use minitests framework for tests https://github.com/seattlerb/minitest
1) with minitests find mechanism to clean-up safely 
2) create convention for naming resources


## currently under WIP

# This are extra tests additionally to the acceptance tests.

The goal of this tests is to be more maintenable and readable then the terraform golang ones.

### How to add tests:

0) create you terraform.tf file.
1) add it to the testsuite.yml file ( order is sequential at moment)
2) run it with `run-testsuite`  (ensure that you builded the plugin with make build and make install before)
