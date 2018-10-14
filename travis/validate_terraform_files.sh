#! /bin/bash
set -x
cd examples
if [ `terraform fmt | wc -c` -ne 0 ]; then echo "terraform files need be formatted! run terraform fmt!"; exit 1; fi
