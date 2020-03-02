#!/bin/bash

#docker pull $3:$4

IP=$1
PASSHRASE='Harbor12345'
notaryServerEndpoint=$5
tag_src=$6
echo $IP

mkdir -p /etc/docker/certs.d/$IP/
mkdir -p ~/.docker/tls/$IP:4443/

cp /notary_ca.crt /etc/docker/certs.d/$IP/
cp /notary_ca.crt ~/.docker/tls/$IP:4443/

mkdir -p ~/.docker/tls/$notaryServerEndpoint/
cp /notary_ca.crt ~/.docker/tls/$notaryServerEndpoint/

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
docker tag $tag_src $IP/$2/$3:$4
docker push $IP/$2/$3:$4
