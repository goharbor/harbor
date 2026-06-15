#!/bin/bash
IP=$1
HARBOR_VERSION=$2
DOCKER_USER=$3
DOCKER_PWD=$4
LOCAL_REGISTRY=$5
LOCAL_REGISTRY_NAMESPACE=$6
make swagger_client
robot -v ip:$IP  -v ip1: -v HARBOR_PASSWORD:Harbor12345 -v DOCKER_USER:$DOCKER_USER -v DOCKER_PWD:$DOCKER_PWD -v HARBOR_ADMIN:admin -v http_get_ca:true -v OPENAPI_GENERATOR_CLI_URL:$OPENAPI_GENERATOR_CLI_URL -v PIP_INDEX_URL:$PIP_INDEX_URL /drone/tests/robot-cases/Group1-Nightly/Setup.robot
cd /drone/tests/robot-cases/Group3-Upgrade
DOCKER_USER=$DOCKER_USER DOCKER_PWD=$DOCKER_PWD  python ./prepare.py -e $IP -v $HARBOR_VERSION -l /drone/tests/apitests/python/ -g $LOCAL_REGISTRY  -p $LOCAL_REGISTRY_NAMESPACE
