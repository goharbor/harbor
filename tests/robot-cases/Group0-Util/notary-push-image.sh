#!/bin/bash

docker pull tomcat:latest

IP=$1
notaryServerEndpoint=$2
PASSHRASE='Harbor12345'

echo $IP
echo "Notary server endpoint: $notaryServerEndpoint"

export DOCKER_CONTENT_TRUST=1
export DOCKER_CONTENT_TRUST_SERVER=https://$notaryServerEndpoint

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
