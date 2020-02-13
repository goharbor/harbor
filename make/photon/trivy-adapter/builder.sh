#!/bin/bash

set +e

if [ -z $1 ]; then
  error "Please set the 'version' variable"
  exit 1
fi

VERSION="$1"

set -e

cd `dirname $0`
cur=$PWD

# the temp folder to store distribution source code...
TEMP=`mktemp -d ${TMPDIR-/tmp}/trivy-adapter.XXXXXX`
git clone https://github.com/aquasecurity/harbor-scanner-trivy.git $TEMP
cd $TEMP; git checkout $VERSION; cd -

echo 'build the trivy adapter binary bases on the golang:1.13.4'
cp Dockerfile.binary $TEMP
docker build -f $TEMP/Dockerfile.binary -t trivy-adapter-golang $TEMP

echo 'copy the trivy adapter binary to local...'
ID=$(docker create trivy-adapter-golang)
docker cp $ID:/go/src/github.com/aquasecurity/harbor-scanner-trivy/scanner-trivy binary

docker rm -f $ID
docker rmi -f trivy-adapter-golang

echo "Build trivy adapter binary success, then to build photon image..."
cd $cur
rm -rf $TEMP
