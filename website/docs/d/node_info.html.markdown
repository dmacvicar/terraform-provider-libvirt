---
layout: "libvirt"
page_title: "Libvirt: libvirt_node_info"
sidebar_current: "docs-libvirt-node-info"
description: |-
  Use this data source to get information about the current node
---

# Data Source: libvirt\_node\_info

Retrieve information about the current node

## Example Usage

```hcl
data "libvirt_node_info" "node" {

}
```

## Argument Reference

This data source has no arguments.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `cpu_cores_per_socket` - Number of CPU cores per each socket
* `cpu_cores_total` - Number of CPU cores in total
* `cpu_model` - CPU Architecture, usually `x86_64`
* `cpu_sockets` - How many CPU sockets are present
* `cpu_threads_per_core` - How many CPU Threads are available per each CPU core
* `memory_total_kb` - The amount of memory installed, in KiB
* `numa_nodes` - How many NUMA nodes/cells are available.
