########################################################################
# libvirt_cluster purpose example
########################################################################

module "pnlpcluster" {
    source = "./modules/libvirt_cluster"

    cluscfg =  {
        uri                   = ["qemu:///system",]
        clusterhosts          = 4
        hostname_format       = "pnlpnode%02d"
   
        vcpu                  = 2
        memory                = 4096
        volumesize            = 42949672960 #17179869184
   
        domain_name           = "pnlplab.com"
        network_range         = "192.168.200.0/24"
        gateway               = "192.168.200.1"
        nameservers           = "192.168.200.1"
        nameserver_forwarder  = "192.168.2.254"
        mtu                   = 9000 # Enable jumbo frames on the network
        start_ip              = 40 
        ip_format             = "192.168.200.%d"
   
        pool_name             = "pnlpzpool"
        pool_path             = "/hpnvmezpool/vms/pnlpcrzpool"
   
        network_name          = "pnlpnet"
   
        clusteruser           = "cg"
        clusteruserfullname   = "C G"
        passwdhash            = ""
        ssh_authorized_key    = "[]"
   
        root_passwd           = "" # Think twice on this one
   
        # Kubernetes stuff, first node is treated as the master node
        pod_network_cidr      = "192.168.210.0/24"
        
    }
}

resource "null_resource" "provision_kube_get" {
    triggers = {
        masternode_ip = module.pnlpcluster.master_node_ip
        #module.pnlpcluster.cluster_nodes[0].name
    }

    provisioner "local-exec" {
        command = "./modules/kube/provision_kube.sh -c get -u ${module.pnlpcluster.cfg.clusteruser} -m  ${module.pnlpcluster.master_node_ip}" 
    }

    provisioner "local-exec" {
        when    = destroy
        command = format("rm -f %s%s %s%s", self.triggers.masternode_ip, "_kube.config", self.triggers.masternode_ip, "_kube_join.cmd")
    }

    provisioner "local-exec" {
        when    = destroy
        command = "sleep 1 && ssh-keygen -f $HOME/.ssh/known_hosts -R ${self.triggers.masternode_ip}" 
    }

}

resource "null_resource" "provision_kube_dashinit" {
    triggers = {
        masternode_ip = module.pnlpcluster.master_node_ip
    }

    provisioner "local-exec" {
        command = "./modules/kube/provision_kube.sh -c dashinit -u ${module.pnlpcluster.cfg.clusteruser} -m  ${module.pnlpcluster.master_node_ip}" 
    }

    depends_on = [null_resource.provision_kube_get]
}

resource "null_resource" "provision_kube_put" {
    count = length(module.pnlpcluster.minion_node_ips)
    triggers = {
        minion_ip = module.pnlpcluster.minion_node_ips[count.index]
    }

    provisioner "local-exec" {
        command = "./modules/kube/provision_kube.sh -c put -u ${module.pnlpcluster.cfg.clusteruser} -m  ${module.pnlpcluster.master_node_ip} -n ${module.pnlpcluster.minion_node_ips[count.index]}" 
    }

    provisioner "local-exec" {
        when    = destroy
        command = "ssh-keygen -f $HOME/.ssh/known_hosts -R ${self.triggers.minion_ip}" 
    }

    depends_on = [null_resource.provision_kube_get]
}

resource "null_resource" "provision_kube_run" {
    count = length(module.pnlpcluster.minion_node_ips)
    triggers = {
        minion_ip = module.pnlpcluster.minion_node_ips[count.index]
    }

    provisioner "local-exec" {
        command = "./modules/kube/provision_kube.sh -c run -u ${module.pnlpcluster.cfg.clusteruser} -m  ${module.pnlpcluster.master_node_ip} -n ${module.pnlpcluster.minion_node_ips[count.index]}" 
    }


    depends_on = [null_resource.provision_kube_put]
}

resource "null_resource" "provision_kube_rookinit" {
    triggers = {
        masternode_ip = module.pnlpcluster.master_node_ip
    }

    provisioner "local-exec" {
        command = "./modules/kube/provision_kube.sh -c rookinit -u ${module.pnlpcluster.cfg.clusteruser} -m  ${module.pnlpcluster.master_node_ip} " 
    }


    depends_on = [null_resource.provision_kube_run]
}

