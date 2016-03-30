#!/usr/bin/env bash

# Run the integration tests with multiple versions of the Docker engine

set -e
set -x

source helpers.bash

if [ `uname` = "Linux" ]; then
	tmpdir_template="$TMPDIR/docker-versions.XXXXX"
else
	# /tmp isn't available for mounting in boot2docker
	tmpdir_template="`pwd`/../../../docker-versions.XXXXX"
fi

tmpdir=`mktemp -d "$tmpdir_template"`
trap "rm -rf $tmpdir" EXIT

if [ "$1" == "-d" ]; then
	start_daemon
fi

# Released versions

versions="1.6.1 1.7.1 1.8.3 1.9.1"

for v in $versions; do
	echo "Extracting Docker $v from dind image"
	binpath="$tmpdir/docker-$v/docker"
	ID=$(docker create dockerswarm/dind:$v)
	docker cp "$ID:/usr/local/bin/docker" "$tmpdir/docker-$v"

	echo "Running tests with Docker $v"
	DOCKER_BINARY="$binpath" DOCKER_VOLUME="$DOCKER_VOLUME" DOCKER_GRAPHDRIVER="$DOCKER_GRAPHDRIVER" ./run.sh

	# Cleanup.
	docker rm -f "$ID"
done

# Latest experimental version

echo "Extracting Docker master from dind image"
binpath="$tmpdir/docker-master/docker"
docker pull dockerswarm/dind-master
ID=$(docker create dockerswarm/dind-master)
docker cp "$ID:/usr/local/bin/docker" "$tmpdir/docker-master"

echo "Running tests with Docker master"
DOCKER_BINARY="$binpath" DOCKER_VOLUME="$DOCKER_VOLUME" ./run.sh

# Cleanup.
docker rm -f "$ID"
