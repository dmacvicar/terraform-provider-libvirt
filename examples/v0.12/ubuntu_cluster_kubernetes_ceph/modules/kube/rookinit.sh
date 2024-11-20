#!/bin/bash

###
# Check on a specific pod status given its name, wait till it 
# is running if a wait time is specified
#
# Returns 0 if pod is running, -1 otherwise
#
# Example Usage: 
#      check_if_running "prometheus-operator" 300
#      echo "Return value is $?"
#      check_if_running "mysql-operator" 29
#      echo "Return value is $?"
###
function check_if_running() {
    retval=-1
    for ((poll=0; poll<$2/10; poll++)) ; do
         status=$(kubectl get pods -A | egrep $1  | tr -s ' ' | cut -d ' ' -f4)
         if [ "$status" == "Running" ] ; then
             echo "$1 is running, moving to next steps"
             retval=0
             break
         else
             echo "$1 is not created or running yet, waiting for 10 seconds"
             #kubectl get pods -A
         fi
         sleep 10
    done

    return $retval
}


ROOKINIT_RETRIES=4

if test -f "$HOME/rook_repo.success"; then
    echo "git repo was already initialized successfully, skipping"
else
   git clone --single-branch --branch release-1.2 https://github.com/rook/rook.git
   retval=$?
   
   if [ ${retval} == 0 ] ; then 
       touch $HOME/rook_repo.success
   else
       echo "git clone of rook repo failed, exiting"
       exit -1
   fi
fi

if test -f "$HOME/rook_stage_one.success"; then
    echo "rook stage one was already initialized successfully, skipping"
    cd rook/cluster/examples/kubernetes/ceph
else
   cd rook/cluster/examples/kubernetes/ceph
   kubectl create -f common.yaml
   kubectl create -f operator.yaml 
   retval=$?
   
   if [ ${retval} == 0 ] ; then 
       if check_if_running "rook-ceph-operator" 300; then
          echo "rook-ceph-operator is now running"
       else
          retval=$?
          echo "Timed out waiting for rook-ceph-operator"
       fi
       touch $HOME/rook_stage_one.success

       ## Wait 2 minutes for operator to do the rest of stuff
       #for ((poll=0;poll<10;poll++)); do
       #   kubectl -n rook-ceph get pods
       #   sleep 10
       #done   
   else
       echo "rook stage two init failed, exiting"
       exit -1
   fi
fi


if test -f "$HOME/rook_stage_two.success"; then
    echo "rook stage two was already initialized successfully, skipping"
else
   cd $HOME/rook/cluster/examples/kubernetes/ceph
   kubectl create -f cluster-test.yaml 
   retval=$?
   
   if [ ${retval} == 0 ] ; then 
       if check_if_running "rook-ceph-mon" 300; then
          echo "rook-ceph-mon is now running"
       else
          retval=$?
          echo "Timed out waiting for rook-ceph-mon"
       fi
       if check_if_running "rook-ceph-mgr" 300; then
          echo "rook-ceph-mgr is now running"
       else
          retval=$?
          echo "Timed out waiting for rook-ceph-mgr"
       fi
       touch $HOME/rook_stage_two.success
       ## Wait 5 minutes for this to complete
       #for ((poll=0;poll<30;poll++)); do
       #   kubectl -n rook-ceph get pods
       #   sleep 10
       #done
   else
       echo "rook stage two init failed, exiting"
       exit -1
   fi
fi


if test -f "$HOME/rook_set_default_storageclass.success"; then
    echo "rook seting as default storageclass was already successfull, skipping"
else
   cd $HOME/rook/cluster/examples/kubernetes/ceph
   kubectl create -f toolbox.yaml 
   kubectl create -f dashboard-external-https.yaml 
   kubectl create -f $HOME/rook/cluster/examples/kubernetes/ceph/csi/rbd/storageclass.yaml 
   kubectl patch storageclass rook-ceph-block -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
   retval=$?
   
   if [ ${retval} == 0 ] ; then 
       touch $HOME/rook_set_default_storageclass.success
   else
       echo "rook setting rook-ceph as the default storageclass failed, exiting"
       exit -1
   fi
fi

if test -f "$HOME/rook_enable_prometheus.success"; then
    echo "rook dashboard was already initialized successfully, skipping"
else
    #kubectl apply -f https://raw.githubusercontent.com/coreos/prometheus-operator/v0.37.0/bundle.yaml
    kubectl apply -f https://raw.githubusercontent.com/coreos/prometheus-operator/v0.36.0/bundle.yaml
    retval=$?
    
    if [ ${retval} == 0 ] ; then 
        if check_if_running "prometheus-operator" 300; then
    	   kubectl create -f $HOME/rook/cluster/examples/kubernetes/ceph/monitoring/service-monitor.yaml
           kubectl create -f $HOME/rook/cluster/examples/kubernetes/ceph/monitoring/prometheus.yaml
           kubectl create -f $HOME/rook/cluster/examples/kubernetes/ceph/monitoring/prometheus-service.yaml
        else
           retval=$?
    	   echo "Timed out waiting for prometheus-operator"
        fi

        if check_if_running "prometheus-rook-prometheus-0" 300; then
            retval=$?
    	    echo "prometheus-rook-prometheus-0 is running"
            touch $HOME/rook_enable_prometheus.success
        else
            retval=$?
    	    echo "Timed out waiting for prometheus-rook-prometheus-0"
        fi
    
        echo "http://$(kubectl -n rook-ceph -o jsonpath={.status.hostIP} get pod prometheus-rook-prometheus-0):30900"
    fi

   if [ ${retval} == 0 ] ; then 
       echo "rook-ceph initialized prometheus operator"
   else
       echo "rook error initializing prometheus, exiting"
       exit -1
   fi
fi

if test -f "$HOME/rook_enable_dashboard.success"; then
    echo "rook dashboard was already initialized successfully, will show again "
    kubectl -n kubernetes-dashboard describe secret $(kubectl -n kubernetes-dashboard get secret | grep admin-user | awk '{print $1}')
    kubectl -n rook-ceph get service rook-ceph-mgr-dashboard-external-https
    kubectl -n rook-ceph get secret rook-ceph-dashboard-password -o jsonpath="{['data']['password']}" | base64 --decode && echo
    echo "http://$(kubectl -n rook-ceph -o jsonpath={.status.hostIP} get pod prometheus-rook-prometheus-0):30900"
 else
    for ((retry=0;retry<${ROOKINIT_RETRIES};retry++)); do
	kubectl -n kubernetes-dashboard describe secret $(kubectl -n kubernetes-dashboard get secret | grep admin-user | awk '{print $1}')
        kubectl -n rook-ceph get service rook-ceph-mgr-dashboard-external-https
        kubectl -n rook-ceph get secret rook-ceph-dashboard-password -o jsonpath="{['data']['password']}" | base64 --decode && echo
        echo "http://$(kubectl -n rook-ceph -o jsonpath={.status.hostIP} get pod prometheus-rook-prometheus-0):30900"
        retval=$?

        if [ ${retval} == 0 ] ; then 
            touch $HOME/rook_enable_dashboard.success
            break
        else
            echo -n .
            sleep 30
            echo "Retrying getting rook-ceph dashbord key, attempt $retry of ${ROOKINIT_RETRIES}"
        fi
    done   
   if [ ${retval} == 0 ] ; then 
       echo "rook-ceph got dashboard secret "
   else
       echo "rook error getting rook-ceph dashboard key, exiting"
       exit -1
   fi
fi


echo -e "rook-ceph was SUCCESSFULLY INITIALIZED"
