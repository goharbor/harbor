#!/bin/bash
set -x

IP=$1
USER=$2
PWD=$3
TARGET=$4
BUNDLE_FILE=$5
DOCKER_USER=$6
DOCKER_PWD=$7
echo $DOCKER_USER
echo $IP

TOKEN=$(curl --user "$DOCKER_USER:$DOCKER_PWD" "https://auth.docker.io/token?service=registry.docker.io&scope=repository:ratelimitpreview/test:pull" | jq -r .token)
curl -v -H "Authorization: Bearer $TOKEN" https://registry-1.docker.io/v2/ratelimitpreview/test/manifests/latest 2>&1 | grep RateLimit

docker login -u $DOCKER_USER -p $DOCKER_PWD
docker login $IP -u $USER -p $PWD

cnab-to-oci fixup  $BUNDLE_FILE --target $TARGET --bundle fixup_bundle.json --auto-update-bundle

TOKEN=$(curl --user "$DOCKER_USER:$DOCKER_PWD" "https://auth.docker.io/token?service=registry.docker.io&scope=repository:ratelimitpreview/test:pull" | jq -r .token)
curl -v -H "Authorization: Bearer $TOKEN" https://registry-1.docker.io/v2/ratelimitpreview/test/manifests/latest 2>&1 | grep RateLimit

cnab-to-oci push  fixup_bundle.json --target $TARGET --auto-update-bundle