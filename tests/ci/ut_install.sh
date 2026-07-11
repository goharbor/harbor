#!/bin/bash
set -x

set -e

sudo apt-get update && sudo apt-get install -y libldap2-dev

if [ -n "$XDG_CONFIG_HOME" ] && [[ "$XDG_CONFIG_HOME" == *'$HOME'* ]]; then
    unset XDG_CONFIG_HOME
fi
if [ -z "$XDG_CONFIG_HOME" ] || [[ "$XDG_CONFIG_HOME" != /* ]]; then
    export XDG_CONFIG_HOME="$HOME/.config"
fi

go env -w GO111MODULE=auto
pwd
GOBIN_DIR="$(go env GOPATH | cut -d: -f1)/bin"
# These are restored from the CI tool cache; only build what is missing.
command -v cover     >/dev/null 2>&1 || go install golang.org/x/tools/cmd/cover@latest
command -v goveralls >/dev/null 2>&1 || go install github.com/mattn/goveralls@latest
command -v misspell  >/dev/null 2>&1 || go install github.com/client9/misspell/cmd/misspell@latest
set -e
# cd ../
# binary will be $(go env GOPATH)/bin/golangci-lint
# go get installation aren't guaranteed to work. We recommend using binary installation.
if ! "${GOBIN_DIR}/golangci-lint" --version 2>/dev/null | grep -q ' 2\.9\.0 '; then
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "${GOBIN_DIR}" v2.9.0
fi
sudo service postgresql stop || echo no postgresql need to be stopped
sleep 2

sudo rm -rf /data/*
sudo -E env "PATH=$PATH" make go_check
sudo ./tests/hostcfg.sh
sudo ./tests/generateCerts.sh
sudo make build -e BUILDTARGET="_build_db _build_registry _build_valkey _build_prepare" -e PULL_BASE_FROM_DOCKERHUB=false -e BUILDREG=true -e BUILDTRIVYADP=true
docker run --rm -v /:/hostfs:z goharbor/prepare:dev gencert -p /etc/harbor/tls/internal
sudo MAKEPATH=$(pwd)/make ./make/prepare
sudo mkdir -p "/data/redis"
sudo mkdir -p /etc/core/ca/ && sudo mv ./tests/ca.crt /etc/core/ca/
sudo mkdir -p /harbor && sudo mv ./VERSION /harbor/UIVERSION
sudo ./tests/testprepare.sh

cd tests && sudo ./ldapprepare.sh && cd ..
env
docker images
sudo sed -i 's/__version__/dev/g' ./make/docker-compose.test.yml
cat ./make/docker-compose.test.yml
sudo mkdir -p ./make/common/config/registry/ && sudo mv ./tests/reg_config.yml ./make/common/config/registry/config.yml
sudo mkdir -p /storage && sudo chown 10000:10000 -R /storage
