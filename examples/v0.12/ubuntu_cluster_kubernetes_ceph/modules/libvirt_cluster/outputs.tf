########################################################################
# Virt Cluster Outputs
########################################################################

# outputs and definitions

output "name" {
  description = "The name of the cluster"
  value       = "${var.cluscfg.domain_name}"
}

output "cfg" {
  description = "The name of the cluster"
  value       = "${var.cluscfg}"
}
output "cluster_network_id" {
  description = "The ID of the cluster's virtual network"
  value       = "${libvirt_network.clusternet.id}"
}

output "cluster_nodes" {
   description = "The list of cluster nodes"
   value = "${libvirt_domain.clusternode}"
}

output "master_node_ip" {
   description = "IP of the master (the first) node in the cluster"
   value = "${format(var.cluscfg.ip_format, var.cluscfg.start_ip)}"
}

output "minion_node_ips" {
   description = "List of IPs of the minion nodes in the cluster"
   value =   [ 
      for num in range(var.cluscfg.clusterhosts -1):
         format(var.cluscfg.ip_format, num + var.cluscfg.start_ip + 1)
    ]

}

