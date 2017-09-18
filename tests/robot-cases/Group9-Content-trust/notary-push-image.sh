#!/bin/bash

docker pull tomcat:latest

IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
PASSHRASE='Harbor12345'

echo $IP

export DOCKER_CONTENT_TRUST=1
export DOCKER_CONTENT_TRUST_SERVER=https://$IP:4443

export NOTARY_ROOT_PASSPHRASE=$PASSHRASE
export NOTARY_TARGETS_PASSPHRASE=$PASSHRASE
export NOTARY_SNAPSHOT_PASSPHRASE=$PASSHRASE
export DOCKER_CONTENT_TRUST_ROOT_PASSPHRASE=$PASSHRASE
export DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE=$PASSHRASE
export DOCKER_CONTENT_TRUST_OFFLINE_PASSPHRASE=$PASSHRASE
export DOCKER_CONTENT_TRUST_TAGGING_PASSPHRASE=$PASSHRASE

docker login -u admin -p Harbor12345 $IP
docker tag tomcat $IP/library/tomcat:latest
docker push $IP/library/tomcat:latest

