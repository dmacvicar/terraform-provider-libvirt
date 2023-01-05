#### CONFIG SECTION ####


variable "cluscfg" {
  description = <<EOF
This is the must provide set of variables to be able to instantiate a  cluster
In the future split out the sub -sections to optiional variables  and sub clusters
like kubernetes
EOF

  default = {
      uri                   = ["qemu:///system",]
      clusterhosts          = 2
      hostname_format       = "vnanode%02d"

      vcpu                  = 1
      memory                = 1024
      volumesize            = 17179869184


      domain_name           = "vnalab.com"
      network_range         = "192.168.200.0/24"
      gateway               = "192.168.200.1"
      nameservers           = "192.168.200.1"
      nameserver_forwarder  = "192.168.2.254"
      mtu                   = 9000 # Enable jumbo frames on the network
      start_ip              = 40 
      ip_format             = "192.168.200.%d"

      pool_name             = "vnapool"
      pool_path             = "/hpnvmezpool/vms/vnapool"

      network_name          = "vnanet"

      clusteruser           = "cg"
      clusteruserfullname   = "C G"
      passwdhash            = "Put Your Password Hash here"
      ssh_authorized_key    = "[ssh-rsa Put your authorized key here ]"

      root_passwd           = "put your password here" # Think twice on this one

      # Kubernetes stuff, first node is treated as the master node
      pod_network_cidr      = "192.168.210.0/24"
     
  }
}

