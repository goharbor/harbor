#!/bin/bash

IP=$1
notaryServerEndpoint=$2
image=$3

docker pull $image:latest

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
docker tag $image $IP/library/$image:latest
docker push $IP/library/$image:latest