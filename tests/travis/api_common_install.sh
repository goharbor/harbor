#!/bin/bash

set +e
sudo rm -fr /data/*
sudo mkdir -p /data

set -e
# prepare cert ...
sudo sed "s/127.0.0.1/$1/" -i tests/generateCerts.sh
sudo ./tests/generateCerts.sh
sudo mkdir -p /etc/docker/certs.d/$1 && sudo cp ./harbor_ca.crt /etc/docker/certs.d/$1/

sudo ./tests/hostcfg.sh

if [ "$2" = 'LDAP' ]; then
    cd tests && sudo ./ldapprepare.sh && cd ..
fi



# prepare a chart file for API_DB test...
sudo curl -o /home/travis/gopath/src/github.com/goharbor/harbor/tests/apitests/python/mariadb-4.3.1.tgz https://storage.googleapis.com/harbor-builds/bin/charts/mariadb-4.3.1.tgz

sudo apt-get update && sudo apt-get install -y --no-install-recommends python-dev openjdk-7-jdk libssl-dev && sudo apt-get autoremove -y && sudo rm -rf /var/lib/apt/lists/*
sudo wget https://bootstrap.pypa.io/get-pip.py && sudo python ./get-pip.py && sudo pip install --ignore-installed urllib3 chardet requests && sudo pip install robotframework==3.0.4 robotframework-httplibrary requests dbbot robotframework-pabot --upgrade
sudo make swagger_client
sudo make install GOBUILDIMAGE=golang:1.12.5 COMPILETAG=compile_golangimage CLARITYIMAGE=goharbor/harbor-clarity-ui-builder:1.6.0 NOTARYFLAG=true CLAIRFLAG=true CHARTFLAG=true
sleep 10
