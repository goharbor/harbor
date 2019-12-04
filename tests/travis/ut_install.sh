#!/bin/bash

set -e

sudo apt-get update && sudo apt-get install -y libldap2-dev
go get -d github.com/docker/distribution
go get -d github.com/docker/libtrust
go get -d github.com/lib/pq
go get golang.org/x/lint/golint
go get github.com/GeertJohan/fgt
go get github.com/dghubble/sling
go get github.com/stretchr/testify
go get golang.org/x/tools/cmd/cover
go get github.com/mattn/goveralls
go get -u github.com/client9/misspell/cmd/misspell
sudo service postgresql stop
sleep 2

sudo -E env "PATH=$PATH" make go_check
sudo ./tests/hostcfg.sh
sudo ./tests/generateCerts.sh
sudo make -f make/photon/Makefile _build_db _build_registry _build_prepare -e VERSIONTAG=dev -e REGISTRYVERSION=${REG_VERSION} -e BASEIMAGETAG=dev
sudo MAKEPATH=$(pwd)/make ./make/prepare
sudo mkdir -p "/data/redis"
sudo mkdir -p /etc/core/ca/ && sudo mv ./tests/ca.crt /etc/core/ca/
sudo mkdir -p /harbor && sudo mv ./VERSION /harbor/UIVERSION
sudo ./tests/testprepare.sh

cd tests && sudo ./ldapprepare.sh && cd ..
sudo sed -i 's/__reg_version__/${REG_VERSION}-dev/g' ./make/docker-compose.test.yml
sudo sed -i 's/__version__/dev/g' ./make/docker-compose.test.yml
sudo mkdir -p ./make/common/config/registry/ && sudo mv ./tests/reg_config.yml ./make/common/config/registry/config.yml
sudo mkdir /storage && sudo chown 10000:10000 -R /storage
