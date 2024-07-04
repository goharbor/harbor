#!/bin/bash

set -ex

sudo apt-get update && sudo apt-get install -y libldap2-dev
sudo go env -w GO111MODULE=auto
pwd
go version
#go get -d github.com/docker/distribution
#go get -d github.com/docker/libtrust
#go get -d github.com/lib/pq
#go get golang.org/x/lint/golint
#go get github.com/GeertJohan/fgt
#go get github.com/dghubble/sling
#set +e
go install golang.org/x/tools/cmd/cover@latest
go install github.com/mattn/goveralls@latest
go install github.com/client9/misspell/cmd/misspell@latest
set -e
sudo service postgresql stop || echo no postgresql need to be stopped
sleep 2

sudo rm -rf /data/*
sudo -E env "PATH=$PATH" make go_check
sudo ./tests/hostcfg.sh
sudo ./tests/generateCerts.sh
sudo make build_base_images -e BASEIMAGETAG=dev
sudo make -f make/photon/Makefile _build_db _build_registry _build_prepare -e VERSIONTAG=dev -e BASEIMAGETAG=dev -e BUILDBIN=true -e REGISTRY_SRC_TAG=v2.7.1
sudo MAKEPATH=$(pwd)/make ./make/prepare
sudo mkdir -p "/data/redis"
sudo mkdir -p /etc/core/ca/ && sudo mv ./tests/ca.crt /etc/core/ca/
sudo mkdir -p /harbor && sudo mv ./VERSION /harbor/UIVERSION
sudo ./tests/testprepare.sh

cd tests && sudo ./ldapprepare.sh && sudo ./admiral.sh && cd ..
env
docker images
sudo sed -i 's/__version__/dev/g' ./make/docker-compose.test.yml
sudo mkdir -p ./make/common/config/registry/ && sudo mv ./tests/reg_config.yml ./make/common/config/registry/config.yml
sudo mkdir /storage && sudo chown 10000:10000 -R /storage
