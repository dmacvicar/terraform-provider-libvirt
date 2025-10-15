# Get list of PCI devices
data "libvirt_node_devices" "pci" {
  capability = "pci"
}

# Get detailed information about a specific PCI device
data "libvirt_node_device_info" "gpu" {
  name = "pci_0000_01_00_0"  # Example: NVIDIA GPU
}

# Access PCI device details
output "gpu_vendor" {
  description = "GPU vendor name"
  value       = data.libvirt_node_device_info.gpu.capability.vendor_name
}

output "gpu_product" {
  description = "GPU product name"
  value       = data.libvirt_node_device_info.gpu.capability.product_name
}

output "gpu_iommu_group" {
  description = "IOMMU group for GPU passthrough"
  value       = data.libvirt_node_device_info.gpu.capability.iommu_group
}

output "gpu_pci_address" {
  description = "PCI address of the GPU"
  value = format("%04x:%02x:%02x.%x",
    data.libvirt_node_device_info.gpu.capability.domain,
    data.libvirt_node_device_info.gpu.capability.bus,
    data.libvirt_node_device_info.gpu.capability.slot,
    data.libvirt_node_device_info.gpu.capability.function
  )
}

# Get network interface details
data "libvirt_node_devices" "net" {
  capability = "net"
}

data "libvirt_node_device_info" "eth0" {
  name = element([for d in data.libvirt_node_devices.net.devices : d if strcontains(d, "eth0")], 0)
}

output "eth0_mac" {
  description = "MAC address of eth0"
  value       = data.libvirt_node_device_info.eth0.capability.address
}

output "eth0_link_state" {
  description = "Link state of eth0"
  value       = data.libvirt_node_device_info.eth0.capability.link_state
}

# Get storage device details
data "libvirt_node_devices" "storage" {
  capability = "storage"
}

data "libvirt_node_device_info" "disk" {
  name = tolist(data.libvirt_node_devices.storage.devices)[0]
}

output "disk_model" {
  description = "Storage device model"
  value       = data.libvirt_node_device_info.disk.capability.model
}

output "disk_serial" {
  description = "Storage device serial number"
  value       = data.libvirt_node_device_info.disk.capability.serial
}

output "disk_size_gb" {
  description = "Storage device capacity in GB"
  value       = data.libvirt_node_device_info.disk.capability.size / 1024 / 1024 / 1024
}
