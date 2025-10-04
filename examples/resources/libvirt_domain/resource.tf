# Basic VM configuration
resource "libvirt_domain" "example" {
  name   = "example-vm"
  memory = 2048
  unit   = "MiB"
  vcpu   = 2
  type   = "kvm"

  os {
    type    = "hvm"
    arch    = "x86_64"
    machine = "q35"
    boot_devices = ["hd", "network"]
  }

  # TODO: Add device configuration once implemented
  # - disks
  # - network interfaces
  # - graphics
}

# VM with UEFI firmware
resource "libvirt_domain" "uefi_example" {
  name   = "uefi-vm"
  memory = 4096
  unit   = "MiB"
  vcpu   = 4
  type   = "kvm"

  os {
    type        = "hvm"
    arch        = "x86_64"
    machine     = "q35"
    firmware    = "efi"
    loader_path = "/usr/share/edk2/x64/OVMF_CODE.secboot.4m.fd"
    loader_readonly = true
    loader_type     = "pflash"
    nvram_path      = "/usr/share/edk2/x64/OVMF_VARS.4m.fd"
    boot_devices    = ["hd"]
  }
}

# VM with direct kernel boot
resource "libvirt_domain" "kernel_boot" {
  name   = "kernel-boot-vm"
  memory = 1024
  unit   = "MiB"
  vcpu   = 1
  type   = "kvm"

  os {
    type        = "hvm"
    arch        = "x86_64"
    kernel      = "/boot/vmlinuz"
    initrd      = "/boot/initrd.img"
    kernel_args = "console=ttyS0 root=/dev/vda1"
  }
}
