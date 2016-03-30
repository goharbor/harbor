#!/bin/sh
set -e
set -x

DOCKER_GRAPHDRIVER=${DOCKER_GRAPHDRIVER:-overlay}
EXEC_DRIVER=${EXEC_DRIVER:-native}

# Set IP address in /etc/hosts for localregistry
IP=$(ifconfig eth0|grep "inet addr:"| cut -d: -f2 | awk '{ print $1}')
echo "$IP localregistry" >> /etc/hosts

sh install_certs.sh localregistry

DOCKER_VERSION=$(docker --version | cut -d ' ' -f3 | cut -d ',' -f1)
major=$(echo "$DOCKER_VERSION"| cut -d '.' -f1)
minor=$(echo "$DOCKER_VERSION"| cut -d '.' -f2)

daemonOpts="daemon"
if [ $major -le 1 ] && [ $minor -lt 9 ]; then
	daemonOpts="--daemon"
fi

docker $daemonOpts --log-level=debug --storage-driver="$DOCKER_GRAPHDRIVER"
