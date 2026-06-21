#!/bin/bash
set -x
set -e

IP=$(hostname -I | awk '{print $1}')
HELLO_WORLD_IMAGE="${HELLO_WORLD_IMAGE:-registry.goharbor.io/dockerhub/library/hello-world:latest}"
HELLO_WORLD_IMAGE_FALLBACK="${HELLO_WORLD_IMAGE_FALLBACK:-docker.io/library/hello-world:latest}"
BUSYBOX_IMAGE="${BUSYBOX_IMAGE:-registry.goharbor.io/dockerhub/library/busybox:latest}"
BUSYBOX_IMAGE_FALLBACK="${BUSYBOX_IMAGE_FALLBACK:-docker.io/library/busybox:latest}"

if ! docker pull "$HELLO_WORLD_IMAGE"; then
    echo "$HELLO_WORLD_IMAGE is unavailable, falling back to $HELLO_WORLD_IMAGE_FALLBACK"
    HELLO_WORLD_IMAGE="$HELLO_WORLD_IMAGE_FALLBACK"
    docker pull "$HELLO_WORLD_IMAGE"
fi
if ! docker pull "$BUSYBOX_IMAGE"; then
    echo "$BUSYBOX_IMAGE is unavailable, falling back to $BUSYBOX_IMAGE_FALLBACK"
    BUSYBOX_IMAGE="$BUSYBOX_IMAGE_FALLBACK"
    docker pull "$BUSYBOX_IMAGE"
fi
docker login -u admin -p Harbor12345 $IP:5000  

docker tag "$HELLO_WORLD_IMAGE" $IP:5000/library/hello-world:latest
docker push $IP:5000/library/hello-world:latest
echo "$? pushed hello world"

docker tag "$BUSYBOX_IMAGE" $IP:5000/library/busybox:latest
docker push $IP:5000/library/busybox:latest
echo "$? pushed busybox"
