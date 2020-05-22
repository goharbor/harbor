#!/bin/bash

IP=$1
USER=$2
PWD=$3
INDEX=$4
IMAGE1=$5
IMAGE2=$6
echo $IP

docker login $IP -u $USER -p $PWD

cat /$HOME/.docker/config.json

sed -i '$d' /$HOME/.docker/config.json
sed -i '$d' /$HOME/.docker/config.json
echo -e "\n        },\n        \"experimental\": \"enabled\"\n}" >> /$HOME/.docker/config.json

cat /$HOME/.docker/config.json

docker manifest create $INDEX $IMAGE1 $IMAGE2
docker manifest push $INDEX
