#!/bin/bash
set -x
set -e

E2E_BASE_TAG=$1
E2E_BASE_IMAGE="${E2E_BASE_IMAGE:-registry.goharbor.io/dockerhub/goharbor/harbor-e2e-engine:$E2E_BASE_TAG}"
E2E_BASE_IMAGE_FALLBACK="${E2E_BASE_IMAGE_FALLBACK:-docker.io/goharbor/harbor-e2e-engine:$E2E_BASE_TAG}"
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
CURRENT_COMMIT=$(git rev-parse --short HEAD)
DOCKER_FILE=Dockerfile
CASE_DIRECTORY=test-files

if ! docker manifest inspect "$E2E_BASE_IMAGE" >/dev/null 2>&1; then
  echo "$E2E_BASE_IMAGE is unavailable, falling back to $E2E_BASE_IMAGE_FALLBACK"
  E2E_BASE_IMAGE="$E2E_BASE_IMAGE_FALLBACK"
fi

if [ -f $DOCKER_FILE  ]; then
  rm -rf $DOCKER_FILE
fi

if [ -d $CASE_DIRECTORY ]; then
  rm -rf $CASE_DIRECTORY
fi

mkdir $CASE_DIRECTORY
mkdir $CASE_DIRECTORY/tests
cp -r ../../tests/{robot-cases,resources,files,apitests} ./$CASE_DIRECTORY/tests
cp -r ../../api ./$CASE_DIRECTORY/
cp ../../Makefile ./$CASE_DIRECTORY/

cat > "$DOCKER_FILE" <<EOF
FROM $E2E_BASE_IMAGE
COPY $CASE_DIRECTORY /drone
EOF

echo "Starting to build image ..."
TARGET_IMAGE=goharbor/harbor-e2e-engine:$CURRENT_BRANCH-$CURRENT_COMMIT
echo "$TARGET_IMAGE"
docker build -t "$TARGET_IMAGE" .
rm -rf $CASE_DIRECTORY
rm -rf $DOCKER_FILE
