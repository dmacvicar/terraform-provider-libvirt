# libvirt_cluster

Create a set of VMS in KVM that are a fully prepped to create a cluster. Example clusters are kubernetes, docker or cloudera 

Provisions the following resources:

* A cluster of VMs and associated network
* A set of VMS
* A common network on which the VMs communicate, with full DNS resolution among VMs
* Static IP assignment for VMS for cluster stability
    * A NAT'd connection to the host/outside world
    * A route from the specified VPC to the peer VPC
    * A new security group on the peer VPC for managing cluster rules
* Optional installation and prep for a Kubernetes cluster
    * Initialization of the kubernetes cluster on the first VM
    * Initialization of Calico for networking
    * Install the dashboard instance and basic credential creation

## Variables

See [variables.tf](variables.tf) for a full listing!

## Outputs

See [outputs.tf](outputs.tf) for a full listing!


## Example

See main.tf in examples for sample usage, includes a full working kubernetes cluster with persistent storage and dashboards for kubernetes, ceph and prometheus 

## Future Improvements

* Make sub-cluster initialization modular. I.e. be able to pick kubernetes, cloudera/hortonworks etc and layer on top of the VMs

