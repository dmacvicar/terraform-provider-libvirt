nohup sudo kubectl --kubeconfig /etc/kubernetes/admin.conf proxy --address=0.0.0.0  --accept-hosts='$'  >prxy.log 2>&1 &
