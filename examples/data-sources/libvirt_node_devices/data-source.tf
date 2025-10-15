# List all devices
data "libvirt_node_devices" "all" {
}

# List only PCI devices (useful for GPU passthrough)
data "libvirt_node_devices" "pci" {
  capability = "pci"
}

# List only network interfaces
data "libvirt_node_devices" "network" {
  capability = "net"
}

# List only USB devices
data "libvirt_node_devices" "usb" {
  capability = "usb_device"
}

# List only storage devices
data "libvirt_node_devices" "storage" {
  capability = "storage"
}

output "all_devices" {
  description = "All devices on the host"
  value       = data.libvirt_node_devices.all.devices
}

output "pci_devices" {
  description = "PCI devices available for passthrough"
  value       = data.libvirt_node_devices.pci.devices
}

output "network_interfaces" {
  description = "Network interfaces on the host"
  value       = data.libvirt_node_devices.network.devices
}
