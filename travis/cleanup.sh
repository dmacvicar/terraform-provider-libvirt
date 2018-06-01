#! /bin/bash

echo "--> Cleanup Domains"
sudo virsh list --all | (grep -E "terraform-test|acceptance-test" || :) | awk '{print $2}' | xargs --no-run-if-empty -n1 -I{} sh -c 'sudo virsh destroy {}; sudo virsh undefine {}'

echo "--> Cleanup Networks"
sudo virsh net-list --all | (grep "acceptance-test-network|test_net" || :) | awk '{print $1}' | xargs --no-run-if-empty -n1 -I{} sh -c 'sudo virsh net-destroy {}; sudo virsh net-undefine {}'

echo "--> Cleanup Volumes"
sudo virsh vol-list --pool default | (grep -E "terraform-acceptance-test*" || :) | awk '{print $1}' | xargs --no-run-if-empty -n1 -I{} sh -c 'sudo virsh vol-delete --pool default {}'
