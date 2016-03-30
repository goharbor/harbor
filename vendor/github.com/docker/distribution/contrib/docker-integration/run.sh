#!/usr/bin/env bash
set -e
set -x

cd "$(dirname "$(readlink -f "$BASH_SOURCE")")"

source helpers.bash

# Root directory of Distribution
DISTRIBUTION_ROOT=$(cd ../..; pwd -P)

volumeMount=""
if [ "$DOCKER_VOLUME" != "" ]; then
	volumeMount="-v ${DOCKER_VOLUME}:/var/lib/docker"
fi

dockerMount=""
if [ "$DOCKER_BINARY" != "" ]; then
	dockerMount="-v ${DOCKER_BINARY}:/usr/local/bin/docker"
else
	DOCKER_BINARY=docker
fi

# Image containing the integration tests environment.
INTEGRATION_IMAGE=${INTEGRATION_IMAGE:-distribution/docker-integration}

if [ "$1" == "-d" ]; then
	start_daemon
	shift
fi

TESTS=${@:-.}

# Make sure we upgrade the integration environment.
docker pull $INTEGRATION_IMAGE

# Start a Docker engine inside a docker container
ID=$(docker run -d -it --privileged $volumeMount $dockerMount \
	-v ${DISTRIBUTION_ROOT}:/go/src/github.com/docker/distribution \
	-e "DOCKER_GRAPHDRIVER=$DOCKER_GRAPHDRIVER" \
	${INTEGRATION_IMAGE} \
	./run_engine.sh)

# Stop container on exit
trap "docker rm -f -v $ID" EXIT


# Wait for it to become reachable.
tries=10
until docker exec "$ID" docker version &> /dev/null; do
	(( tries-- ))
	if [ $tries -le 0 ]; then
		echo >&2 "error: daemon failed to start"
		exit 1
	fi
	sleep 1
done

# If no volume is specified, transfer images into the container from
# the outer docker instance
if [ "$DOCKER_VOLUME" == "" ]; then
	# Make sure we have images outside the container, to transfer to the container.
	# Not much will happen here if the images are already present.
	docker-compose pull
	docker-compose build

	# Transfer images to the inner container.
	for image in "$INTEGRATION_IMAGE" registry:0.9.1 dockerintegration_nginx dockerintegration_registryv2; do
		docker save "$image" | docker exec -i "$ID" docker load
	done
fi

# Run the tests.
docker exec -it "$ID" sh -c "./test_runner.sh $TESTS"

