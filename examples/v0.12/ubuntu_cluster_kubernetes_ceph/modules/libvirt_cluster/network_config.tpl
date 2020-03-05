version: 2
ethernets:
        ens3:
                match:
                        driver: virtio_net
                addresses:
                        - ${ip}/24
                gateway4: ${gateway}
                nameservers: 
                         addresses: [${nameservers}]
                         search: [${searchdomain}]
                mtu: ${mtu}

