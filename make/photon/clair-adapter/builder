#!/bin/bash

set +e

if [ -z $1 ]; then
  error "Please set the 'version' variable"
  exit 1
fi

VERSION="$1"

set -e

# the temp folder to store binary file...
mkdir -p binary
rm -rf binary/harbor-scanner-clair || true

cd `dirname $0`
cur=$PWD

# the temp folder to store distribution source code...
TEMP=`mktemp -d ${TMPDIR-/tmp}/clair-adapter.XXXXXX`
git clone https://github.com/goharbor/harbor-scanner-clair.git $TEMP
cd $TEMP; git checkout $VERSION; cd -

echo 'build the clair adapter binary bases on the golang:1.13.4'
cp Dockerfile.binary $TEMP
docker build -f $TEMP/Dockerfile.binary -t clair-adapter-golang $TEMP

echo 'copy the clair adapter binary to local...'
ID=$(docker create clair-adapter-golang)
docker cp $ID:/go/src/github.com/goharbor/harbor-scanner-clair/harbor-scanner-clair binary

docker rm -f $ID
docker rmi -f clair-adapter-golang

echo "Build clair adapter binary success, then to build photon image..."
cd $cur
rm -rf $TEMP
