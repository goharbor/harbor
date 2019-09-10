#!/usr/bin/env bash

[[ -z "${DEBUG:-}" ]] || set -x

mkdir -p /home/travis/go/src/github.com/goharbor/harbor
ln -s /home/travis/go /home/travis/gopath
cp -R /h/* /home/travis/go/src/github.com/goharbor/harbor/

if [[ -d /home/travis/go/src/github.com/goharbor/harbor/src/portal/node-modules ]]; then
  sudo rm -rf /home/travis/go/src/github.com/goharbor/harbor/src/portal/node-modules
fi

# shellcheck disable=SC1090
source "$(cd "$(dirname "${BASH_SOURCE[0]}")"; pwd)/.env"

if ! grep -q insecure-registry /etc/default/docker; then
  sudo sed -i '$a DOCKER_OPTS=\"--insecure-registry '"$IP"':5000\"' /etc/default/docker
  sudo service docker start
fi
