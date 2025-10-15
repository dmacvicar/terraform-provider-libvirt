data "libvirt_node_info" "host" {
}

output "host_cpu_model" {
  description = "CPU model of the host system"
  value       = data.libvirt_node_info.host.cpu_model
}

output "host_memory_gb" {
  description = "Total host memory in GB"
  value       = data.libvirt_node_info.host.memory_total_kb / 1024 / 1024
}

output "host_cpu_topology" {
  description = "CPU topology of the host"
  value = {
    sockets        = data.libvirt_node_info.host.cpu_sockets
    cores_per_socket = data.libvirt_node_info.host.cpu_cores_per_socket
    threads_per_core = data.libvirt_node_info.host.cpu_threads_per_core
    total_cores    = data.libvirt_node_info.host.cpu_cores_total
  }
}
