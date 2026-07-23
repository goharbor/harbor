#!/bin/bash

set +e

if [ -z $1 ]; then
  echo "Please set the 'version' variable"
  exit 1
fi

if [ -z $2 ]; then
  echo "Please set the 'trivy_src' variable"
  exit 1
fi

VERSION="$1"
TRIVY_SRC="$2"
GOBUILDIMAGE="$3"
DOCKERNETWORK="$4"
# Trivy stamps its version without the leading "v" (see version() in its
# magefiles/magefile.go), so `trivy --version` reads e.g. 0.71.1, not v0.71.1.
VERSION_NO_V="${VERSION#v}"

set -e

cd $(dirname $0)
cur=$PWD

# The temporary directory to clone the Trivy scanner source code. We build the
# binary from the git tag instead of downloading the release tarball: Aqua prunes
# and yanks old release assets (the tarball 404s for aged-out versions), whereas
# the git tags are never removed, so building from source keeps older Harbor tags
# buildable.
# Shallow, single-branch clone of just the tag: the Trivy repo is large and we
# stamp the version explicitly (below), so the full history is not needed.
TEMP=$(mktemp -d ${TMPDIR-/tmp}/trivy.XXXXXX)
git clone --depth 1 --single-branch -b $VERSION $TRIVY_SRC $TEMP

echo "Building Trivy scanner binary $VERSION ..."
cp Dockerfile.trivy-binary $TEMP
docker build --network=$DOCKERNETWORK --build-arg golang_image=$GOBUILDIMAGE --build-arg trivy_version=$VERSION_NO_V -f $TEMP/Dockerfile.trivy-binary -t trivy-golang $TEMP

echo "Copying Trivy scanner binary from the container to the local directory..."
ID=$(docker create trivy-golang)
docker cp $ID:/go/src/github.com/aquasecurity/trivy/trivy binary/trivy

docker rm -f $ID
docker rmi -f trivy-golang

echo "Building Trivy scanner binary finished successfully"
cd $cur
rm -rf $TEMP
