#cloud-config
# vim: syntax=yaml
#

hostname: ${hostname}
ssh_pwauth: True
chpasswd:
  list: |
     root:${root_passwd}
  expire: False

users:
    - default
    - name: ${clusteruser}
      gecos: ${clusteruserfullname}
      primary_group: ${clusteruser}
      groups: users,sudo
      shell: /bin/bash
      lock_passwd: false
      sudo: ALL=(ALL) NOPASSWD:ALL
      passwd: ${passwdhash} 
      ssh_authorized_keys: ${ssh_authorized_key}

ssh_authorized_keys: ${ssh_authorized_key}

 
apt:
    sources:
      kubesource1: 
       source: "deb http://apt.kubernetes.io/ kubernetes-xenial main"
       key: |
         -----BEGIN PGP PUBLIC KEY BLOCK-----
    
         mQENBFrBaNsBCADrF18KCbsZlo4NjAvVecTBCnp6WcBQJ5oSh7+E98jX9YznUCrN
         rgmeCcCMUvTDRDxfTaDJybaHugfba43nqhkbNpJ47YXsIa+YL6eEE9emSmQtjrSW
         IiY+2YJYwsDgsgckF3duqkb02OdBQlh6IbHPoXB6H//b1PgZYsomB+841XW1LSJP
         YlYbIrWfwDfQvtkFQI90r6NknVTQlpqQh5GLNWNYqRNrGQPmsB+NrUYrkl1nUt1L
         RGu+rCe4bSaSmNbwKMQKkROE4kTiB72DPk7zH4Lm0uo0YFFWG4qsMIuqEihJ/9KN
         X8GYBr+tWgyLooLlsdK3l+4dVqd8cjkJM1ExABEBAAG0QEdvb2dsZSBDbG91ZCBQ
         YWNrYWdlcyBBdXRvbWF0aWMgU2lnbmluZyBLZXkgPGdjLXRlYW1AZ29vZ2xlLmNv
         bT6JATgEEwECACwFAlrBaNsJEGoDCyG6B/T7AhsPBQkFo5qABgsJCAcDAgYVCAIJ
         CgsEFgIDAQAAJr4IAM5lgJ2CTkTRu2iw+tFwb90viLR6W0N1CiSPUwi1gjEKMr5r
         0aimBi6FXiHTuX7RIldSNynkypkZrNAmTMM8SU+sri7R68CFTpSgAvW8qlnlv2iw
         rEApd/UxxzjYaq8ANcpWAOpDsHeDGYLCEmXOhu8LmmpY4QqBuOCM40kuTDRd52PC
         JE6b0V1t5zUqdKeKZCPQPhsS/9rdYP9yEEGdsx0V/Vt3C8hjv4Uwgl8Fa3s/4ag6
         lgIf+4SlkBAdfl/MTuXu/aOhAWQih444igB+rvFaDYIhYosVhCxP4EUAfGZk+qfo
         2mCY3w1pte31My+vVNceEZSUpMetSfwit3QA8EE=
         =csu4
         -----END PGP PUBLIC KEY BLOCK-----
##
package_update: true
package_upgrade: true

packages:
    # This package block is optional, enable only if you want to run a kubernetes cluster
    - kubelet
    - kubeadm
    - kubectl
    - docker.io
    - curl
    - apt-transport-https 
    - ca-certificates 
    - software-properties-common 
    - gnupg2
    # This package block is optional for general system management 
    - neofetch
    - htop
    - iperf
    # This is a must install package 
    - qemu-guest-agent


# Note: Content strings here are truncated for example purposes.
#write_files:
${write_files_section}

runcmd:
  # This run block is optional, enable only if you want to run a kubernetes cluster
  - [ swapoff, -a ]
  - [ apt-mark, hold, kubeadm, kubelet, kubectl ]
  - [ systemctl, daemon-reload ]
  - [ systemctl, enable, docker ]
  - [ systemctl, start, --no-block, docker ]
  - [ kubeadm,  config, images, pull ]
  ${kube_master_init}

#phone_home:
#    url: http://example.com/$INSTANCE_ID/
#    post:
#        - pub_key_dsa
#        - instance_id
#        - fqdn
#    tries: 10
final_message: "Successfully brought up cluster node: ${hostname}"
