#!/bin/bash
set -x
set -e

IP=$1
USER=$2
PWD=$3
INDEX=$4
IMAGE1=$5
IMAGE2=$6

docker login $IP -u $USER -p $PWD
docker manifest create $INDEX $IMAGE1 $IMAGE2
docker manifest push $INDEX
