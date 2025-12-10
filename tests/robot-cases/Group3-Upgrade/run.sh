#!/bin/bash
IP=$1
HARBOR_VERSION=$2
DOCKER_USER=$3
DOCKER_PWD=$4
LOCAL_REGISTRY=$5
LOCAL_REGISTRY_NAMESPACE=$6
# if $7 exists, set skip_proxycache(-s) to true
if [ -n "$7" ]; then
  SKIP_PROXYCACHE="-s"
else
  SKIP_PROXYCACHE=""
fi
make swagger_client
robot -v ip:$IP  -v ip1: -v HARBOR_PASSWORD:Harbor12345 -v DOCKER_USER:$DOCKER_USER -v DOCKER_PWD:$DOCKER_PWD -v http_get_ca:true -v HARBOR_ADMIN:admin /drone/tests/robot-cases/Group1-Nightly/Setup.robot
cd /drone/tests/robot-cases/Group3-Upgrade
DOCKER_USER=$DOCKER_USER DOCKER_PWD=$DOCKER_PWD  python ./prepare.py -e $IP -v $HARBOR_VERSION -l /drone/tests/apitests/python/ -g $LOCAL_REGISTRY  -p $LOCAL_REGISTRY_NAMESPACE $SKIP_PROXYCACHE
