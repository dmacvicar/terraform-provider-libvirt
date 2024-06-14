#!/bin/bash

# Inputs are clusteruser, masternode ip,  minion node ips
# provision_kube.sh -c get -u cg -m 192.168.200.40, 
# provision_kube.sh -c put -u cg -m 192.168.200.40 -n 192.168.200.40
# provision_kube.sh -c run -u cg -n 192.168.200.40
# provision_kube.sh -c dashinit -u cg -m 192.168.200.41

HELP_TEXT="\n
USAGE:\n
-c: command (one of get, put, run or dashinit)\n
-h: help\n
-m: master node ip\n
-n: minion node ip/s\n
-u: username for ssh/scp\n
"
HELP_EXAMPLES="Examples:\n
provision_kube.sh -c get -u cg -m 192.168.200.40,\n
provision_kube.sh -c put -u cg -m 192.168.200.40 -n 192.168.200.40\n
provision_kube.sh -c run -u cg -n 192.168.200.40\n
provision_kube.sh -c dashinit -u cg -m 192.168.200.40\n
provision_kube.sh -r rookinit -u cg -m 192.168.200.40\n
"
HELP_CMDS="\n
get: \tGets kubernetes admin configuration file and the join command for\n
\tnodes to add themselves  from the master node\n
put: \tDistributes the cluster join command to one or more nodes\n
run: \tExecutes the cluster join command on one or more nodes\n
dashinit: \tDeploys the dashboard to master node, creates a\n
\tservice account and cluster binds it to the dashboard
rookinit: \tDeploys rook-ceph cluster and sets it as the default storage class , a\n
\tenables toolbox and ceph dashboard
\n
"
 
BINLOCATION="/usr/local/bin"
SCPBINARY="scp"
SSHBINARY="ssh"
SCPOPTIONS="-o StrictHostKeyChecking=no -o CheckHostIP=no -o ConnectTimeout=30 -o BatchMode=yes"
INIT_WAIT_TIME=2
RETRY_WAIT_TIME=30
GET_RETRIES=20
PUT_RETRIES=3
RUN_RETRIES=2
DASH_RETRIES=2
ROOK_RETRIES=2

#
while getopts c:u:m:n:h option
do
case "${option}"
in
c) CMD=${OPTARG};;
u) CLUSTERUSER=${OPTARG};;
m) MASTERIP=${OPTARG};;
n) NODEIPS=${OPTARG};;
r) MASTERIP=${OPTARG};;
h) HELP=${OPTARG};;
esac
done

REMOTE_CONFIG="/home/${CLUSTERUSER}/.kube/config"
REMOTE_JOIN_CMD=${BINLOCATION}/"kube_join_cmd.sh"
GET_CMD_FILES="$REMOTE_CONFIG $REMOTE_JOIN_CMD"

RECV_CONFIG_FILENAME="./${MASTERIP}_kube.config"
RECV_JOINNODE_FILENAME="./${MASTERIP}_kube_join.cmd"
GET_CMD_RECV_FILES="$RECV_CONFIG_FILENAME $RECV_JOINNODE_FILENAME"
PUT_JOINNODE_FILENAME="/home/${CLUSTERUSER}/kube_join_cmd.sh"

NODE_REMOTE_JOIN_CMD="sudo /home/${CLUSTERUSER}/kube_join_cmd.sh"

DASH_FILES='modules/kube/dash_createsa.yaml modules/kube/dash_rolebind.yaml modules/kube/startproxy.sh'

DASH_DEPLOY_CMD="sudo kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.0.0-rc5/aio/deploy/recommended.yaml"
DASH_CREATESA_CMD="sudo kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f /home/${CLUSTERUSER}/dash_createsa.yaml"
DASH_ROLEBIND_CMD="sudo kubectl --kubeconfig /etc/kubernetes/admin.conf apply -f /home/${CLUSTERUSER}/dash_rolebind.yaml"

PROXY_CMD="./startproxy.sh" 

ROOK_LOCAL_CMD_FILENAME="modules/kube/rookinit.sh"
ROOK_REMOTE_CMD_FILENAME="/home/${CLUSTERUSER}/rookinit.sh"
ROOK_INIT_CMD="${ROOK_REMOTE_CMD_FILENAME}"

# debug
if [ "${CMD}" == "get" ] ; then
    echo "Waiting for ${INIT_WAIT_TIME} seconds for master to initialize"
    sleep ${INIT_WAIT_TIME} 

    for ((retry=0;retry<${GET_RETRIES};retry++)); do
        # scp the kube_join cmd back to terraform host 
        $SCPBINARY $SCPOPTIONS $CLUSTERUSER@${MASTERIP}:${REMOTE_CONFIG} ${RECV_CONFIG_FILENAME} 

        retval=$?

        if [ ${retval} == 0 ] ; then 
            break
        else
            if [ "$retry" != "${GET_RETRIES}" ] ; then
		    for ((wait=0;wait< ${RETRY_WAIT_TIME};wait++)); do 
                        #echo -n .  # terraform puts it on a new line anyways
            	        sleep 1
                    done
                    echo "Retrying config and join cmd scp, attempt $retry of ${GET_RETRIES}"
 	    fi
        fi
    done
    echo 
    if [ ${retval} -ne 0 ]; then
        echo "Could not get cluster config and join command from master, exiting with rc ${retval}"
        exit $retval
    fi
    echo
    
    for ((retry=0;retry<${GET_RETRIES};retry++)); do
        $SCPBINARY $SCPOPTIONS $CLUSTERUSER@${MASTERIP}:${REMOTE_JOIN_CMD} ${RECV_JOINNODE_FILENAME}

        retval=$?

        if [ ${retval} == 0 ] ; then 
            chmod u+x  ${RECV_JOINNODE_FILENAME}
            break
        else
            echo "retried join cmd scp, attempt $retry of ${GET_RETRIES} "
            if [ "$retry" != "${GET_RETRIES}" ] ; then
        	    for ((wait=0;wait< ${RETRY_WAIT_TIME};wait++)); do 
                        echo -n .
            	        sleep 1
                    done
            fi
        fi
    done
    echo 
    
    if [ ${retval} -ne 0 ]; then
        echo "Could not get join command file from master ${MASTERIP}, exiting with rc ${retval}"
        exit $retval
    fi

elif [ "${CMD}" == "put" ] ; then
    # distribute the run file to the minions 
    for node in  ${NODEIPS}; do
        echo "Distributing run file to ${node}"
        echo  $SCPBINARY $SCPOPTIONS ${RECV_JOINNODE_FILENAME} $CLUSTERUSER@${node}:${PUT_JOINNODE_FILENAME} 
        for (( i = 1; i <= ${PUT_RETRIES}; i++ )); do 
            if $SCPBINARY $SCPOPTIONS ${RECV_JOINNODE_FILENAME} $CLUSTERUSER@${node}:${PUT_JOINNODE_FILENAME} ; then
                retval=$?
                break
            else
                retval=$?
                echo "retried distribution to ${node}, attempt ${i} of ${PUT_RETRIES} "
            fi
        done
        if [ ${retval} -ne 0 ]; then
            echo "Could not distribute run file to ${node}, exiting with rc ${retval}"
            exit $retval
        fi
    done

elif [ "${CMD}" == "run" ] ; then
    # execute the join node command on the minion
    for node in  ${NODEIPS}; do
        echo "Executing join cluster on node ${node}"
        for (( i = 1; i <= ${RUN_RETRIES}; i++ )); do 
	    #debug 
            if $SSHBINARY $SCPOPTIONS $CLUSTERUSER@${node} ${NODE_REMOTE_JOIN_CMD}  ; then
                retval=$?
                break
            else
                retval=$?
                echo "retried joining ${node} to cluster, attempt ${i} of ${RUN_RETRIES} "
            fi
        done
        if [ ${retval} -ne 0 ]; then
            echo "Could not join ${node}, exiting with rc ${retval}"
            exit $retval
        fi
    done

elif [ "${CMD}" == "dashinit" ] ; then
    echo "Executing dashboard init on master "
    for ((retry=0;retry<${DASH_RETRIES};retry++)); do
        # scp the yaml files 
        $SCPBINARY $SCPOPTIONS $DASH_FILES $CLUSTERUSER@${MASTERIP}:

        retval=$?

        if [ ${retval} == 0 ] ; then 
            break
        else
            echo "retried config scp, attempt $retry of ${DASH_RETRIES} "
        fi
    done
    echo 
    if [ ${retval} -ne 0 ]; then
        echo "Could not upload dashboard yaml files to master, exiting with rc ${retval}"
        exit $retval
    fi

    echo "Deploying dashboard on master "
    for (( i = 1; i <= ${DASH_RETRIES}; i++ )); do 
        if $SSHBINARY $SCPOPTIONS $CLUSTERUSER@${MASTERIP} ${DASH_DEPLOY_CMD}  ; then
            retval=$?
            break
        else
            retval=$?
            echo "retried deploying dashboard service account,  attempt ${i} of ${DASH_RETRIES} "
        fi
    done
    if [ ${retval} -ne 0 ]; then
        echo "Could not deploy dashboard service account, exiting with rc ${retval}"
        exit $retval
    fi

    for (( i = 1; i <= ${DASH_RETRIES}; i++ )); do 
        if $SSHBINARY $SCPOPTIONS $CLUSTERUSER@${MASTERIP} ${DASH_CREATESA_CMD}  ; then
            retval=$?
            break
        else
            retval=$?
            echo "retried creating dashboard service account,  attempt ${i} of ${DASH_RETRIES} "
        fi
    done
    if [ ${retval} -ne 0 ]; then
        echo "Could not create dashboard service account, exiting with rc ${retval}"
        exit $retval
    fi

    echo "Executing dashboard rolebind on master "
    for (( i = 1; i <= ${DASH_RETRIES}; i++ )); do 
        if $SSHBINARY $SCPOPTIONS $CLUSTERUSER@${MASTERIP} ${DASH_ROLEBIND_CMD}  ; then
            retval=$?
            break
        else
            retval=$?
            echo "retried rolebinding dashboard,  attempt ${i} of ${DASH_RETRIES} "
        fi
    done
    if [ ${retval} -ne 0 ]; then
        echo "Could not cluster rolebind dashboard, exiting with rc ${retval}"
        exit $retval
    fi

    echo "Starting proxy on master "
    for (( i = 1; i <= ${DASH_RETRIES}; i++ )); do 
        #echo $SSHBINARY $SCPOPTIONS $CLUSTERUSER@${MASTERIP} $PROXY_CMD
        if $SSHBINARY $SCPOPTIONS $CLUSTERUSER@${MASTERIP} $PROXY_CMD  ; then
            retval=$?
            break
        else
            retval=$?
            echo "retried starting proxy,  attempt ${i} of ${DASH_RETRIES} "
        fi
    done
    if [ ${retval} -ne 0 ]; then
        echo "Could not start proxy, exiting with rc ${retval}"
        exit $retval
    fi
# ROOK init section
elif [ "${CMD}" == "rookinit" ] ; then
    echo "Executing rook-ceph init on master "
    for ((retry=0;retry<${ROOK_RETRIES};retry++)); do
        # scp the yaml files 
        $SCPBINARY $SCPOPTIONS $ROOK_LOCAL_CMD_FILENAME $CLUSTERUSER@${MASTERIP}:

        retval=$?

        if [ ${retval} == 0 ] ; then 
            break
        else
            echo "retried rookinit scp, attempt $retry of ${ROOK_RETRIES} "
        fi
    done
    echo 
    echo "Initializing rook-ceph on master "
    for (( i = 1; i <= ${ROOK_RETRIES}; i++ )); do 
        if $SSHBINARY $SCPOPTIONS $CLUSTERUSER@${MASTERIP} ${ROOK_INIT_CMD}  ; then
            retval=$?
            break
        else
            retval=$?
            echo "retried initializing rook-ceph on master,  attempt ${i} of ${ROOK_RETRIES} "
        fi
    done
    if [ ${retval} -ne 0 ]; then
        echo "Could not successfully initilize rook-ceph , exiting with rc ${retval}"
        exit $retval
    fi

# HELP SECTION
elif [ "${CMD}" == "help" ] ; then
     echo -e ${HELP_TEXT}
     echo -e ${HELP_CMDS}
     echo -e ${HELP_EXAMPLES}

else
    echo "Unknown or no command specified ${CMD}"
     echo -e ${HELP_TEXT}

fi
exit 0
