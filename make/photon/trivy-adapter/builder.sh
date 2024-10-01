#!/bin/bash

set -e

if [ -z $1 ]; then
  echo "Please set the 'version' variable"
  exit 1
fi

VERSION="$1"

cd $(dirname $0)

# The temporary directory to clone Trivy adapter source code
TEMP=$(mktemp -d ${TMPDIR-/tmp}/trivy-adapter.XXXXXX)
git clone -b $VERSION --depth 1 https://github.com/aquasecurity/harbor-scanner-trivy.git $TEMP

echo "Building Trivy adapter binary based on golang:1.22.3..."
DOCKER_BUILDKIT=1 docker build -f Dockerfile.binary -o binary/ $TEMP

echo "Building Trivy adapter binary finished successfully"
rm -rf $TEMP
