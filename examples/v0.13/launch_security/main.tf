# Configure the provider
provider "libvirt" {
  uri = "qemu:///system"
}

# SEV example
resource "libvirt_domain" "sev_example" {
  name   = "sev-example"
  memory = 2048
  vcpu   = 2

  launch_security {
    type             = "sev"
    cbitpos          = 47
    reduced_phys_bits = 1
    policy           = 3
  }
}

# SEV-SNP example
resource "libvirt_domain" "sev_snp_example" {
  name   = "sev-snp-example"
  memory = 2048
  vcpu   = 2

  launch_security {
    type             = "sev-snp"
    cbitpos          = 51
    reduced_phys_bits = 1
    policy           = 7
    kernel_hashes    = "enabled"
  }
}

# S390-PV example
resource "libvirt_domain" "s390_pv_example" {
  name   = "s390-pv-example"
  memory = 2048
  vcpu   = 2

  launch_security {
    type = "s390-pv"
  }
}