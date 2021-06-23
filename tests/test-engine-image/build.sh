#!/bin/bash
set -x
set -e

IMAGE_FOR=$1
VERSION=$2

CMD_BASE="cat Dockerfile.common"
SRC_FILE=""
DST_FILE=Dockerfile

echo "Starting to prepare Dockerfile for $IMAGE_FOR ..."
if [ "$IMAGE_FOR" == "api" ]; then
    SRC_FILE=Dockerfile.api_test
else
    SRC_FILE=Dockerfile.ui_test
fi

if [ ! -r $SRC_FILE ]; then
    echo "File $SRC_FILE does not exists at all!"
    exit -1
fi

if [ -f $DST_FILE  ]; then
    rm $DST_FILE
fi
$CMD_BASE $SRC_FILE >> $DST_FILE

echo "Starting to build image ..."
TARGET_IMAGE=goharbor/harbor-e2e-engine:${VERSION}-${IMAGE_FOR}
docker build  -t $TARGET_IMAGE .

