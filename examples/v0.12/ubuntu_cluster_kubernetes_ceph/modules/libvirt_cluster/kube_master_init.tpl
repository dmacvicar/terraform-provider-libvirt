write_files:
####
### Create kubernetes cluster script
####
    -   content: |
            #!/bin/bash
            
            sudo kubeadm init --pod-network-cidr=${pod_network_cidr}
            
            sleep 20 

            sudo kubeadm token create --print-join-command > /usr/local/bin/kube_join_cmd.sh

            mkdir -p /home/${clusteruser}/.kube
            sudo cp -i /etc/kubernetes/admin.conf /home/${clusteruser}/.kube/config
            sudo cp -i /usr/local/bin/kube_join_cmd.sh /home/${clusteruser}/.kube/
            sudo chown -R ${clusteruser}:${clusteruser} /home/${clusteruser}/.kube
            
            kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f https://docs.projectcalico.org/manifests/calico.yaml

   
        owner: root:root

        path: /usr/local/bin/kube_master_init.sh
        permissions: '0755'

####
### Monitor pods and nodes
####
    -   content: |
            #!/bin/bash
            watch -n 1 -d kubectl get pods -o wide  --all-namespaces

        owner: root:root

        path: /usr/local/bin/monitor_pods.sh

        permissions: '0755'

####
### Dashboard token script
####
    -   content: |
            #!/bin/bash
            kubectl -n kubernetes-dashboard describe secret $(kubectl -n kubernetes-dashboard get secret | grep admin-user | awk '{print $1}')

        #owner: root:root

        path: /usr/local/bin/dashboard_token.sh

        permissions: '0755'

