#!/bin/bash

#docker pull $3:$4
set -x

IP=$1
notaryServerEndpoint=$5
tag_src=$6
USER=$7
PASSHRASE=$8
echo $IP

# mkdir -p /etc/docker/certs.d/$IP/
# cp /drone/harbor_ca.crt /etc/docker/certs.d/$IP/
# mkdir -p ~/.docker/tls/$notaryServerEndpoint/
# cp /drone/harbor_ca.crt ~/.docker/tls/$notaryServerEndpoint/

# mkdir -p /etc/docker/certs.d/$IP/
# cp /notary_ca.crt /etc/docker/certs.d/$IP/
# mkdir -p ~/.docker/tls/$notaryServerEndpoint/
# cp /notary_ca.crt ~/.docker/tls/$notaryServerEndpoint/

# cp /etc/docker/certs.d/$IP/harbor_ca.crt /etc/docker/certs.d/$IP/notary_ca.crt
# cp ~/.docker/tls/$notaryServerEndpoint/harbor_ca.crt ~/.docker/tls/$notaryServerEndpoint/notary_ca.crt
# cat ~/.docker/tls/$notaryServerEndpoint/harbor_ca.crt
# cat ~/.docker/tls/$notaryServerEndpoint/notary_ca.crt
# cat /etc/docker/certs.d/$IP/harbor_ca.crt
# cat /etc/docker/certs.d/$IP/notary_ca.crt

export DOCKER_CONTENT_TRUST=1
export DOCKER_CONTENT_TRUST_SERVER=https://$notaryServerEndpoint

export NOTARY_ROOT_PASSPHRASE=$PASSHRASE
export NOTARY_TARGETS_PASSPHRASE=$PASSHRASE
export NOTARY_SNAPSHOT_PASSPHRASE=$PASSHRASE
export DOCKER_CONTENT_TRUST_ROOT_PASSPHRASE=$PASSHRASE
export DOCKER_CONTENT_TRUST_REPOSITORY_PASSPHRASE=$PASSHRASE
export DOCKER_CONTENT_TRUST_OFFLINE_PASSPHRASE=$PASSHRASE
export DOCKER_CONTENT_TRUST_TAGGING_PASSPHRASE=$PASSHRASE

docker login -u $USER -p $PASSHRASE $IP
docker tag $tag_src $IP/$2/$3:$4
docker push $IP/$2/$3:$4
