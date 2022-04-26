#!/bin/bash

set -x
set +e
sudo rm -fr /data/*
sudo mkdir -p /data
DIR="$(cd "$(dirname "$0")" && pwd)"

set -e
# prepare cert ...
sudo sed "s/127.0.0.1/$1/" -i tests/generateCerts.sh
sudo ./tests/generateCerts.sh
sudo mkdir -p /etc/docker/certs.d/$1 && sudo cp ./tests/harbor_ca.crt $DIR/../../tests/ca.crt && sudo cp ./tests/harbor_ca.crt /etc/docker/certs.d/$1/ && rm -rf ~/.docker/ &&  mkdir -p ~/.docker/tls/$1:4443/ && sudo cp ./tests/harbor_ca.crt ~/.docker/tls/$1:4443/
ls -l $DIR/../../tests

sudo ./tests/hostcfg.sh

#---------------Set DNS for docker v20--------------------------#
# In docker v20, it fixed an issue named  "Wrong resolv.conf    #
# used on Ubuntu 19", this fix caused DNS solve problem         #
# in container. So the current work round is read DNS server    #
# from system and set the value in /etc/docker/daemon.json.     #
#                                                               #
# Note: In LDAP pipeline, this setting must be done before      #
# LDAP prepare phase, since LDAP service is a docker service.   #

ip addr
dns_ip=$(netplan ip leases eth0 | grep -i dns | awk -F = '{print $2}')
dns_ip_list=$(echo $dns_ip | tr " " "\n")
dns_cfg=""
for ip in $dns_ip_list
do
    dns_cfg="$dns_cfg,\"$ip\""
done

cat /etc/docker/daemon.json

if [ $(cat /etc/docker/daemon.json |grep \"dns\" |wc -l) -eq 0 ];then
    sudo sed "s/}/,\n   \"dns\": [${dns_cfg:1}]\n}/" -i /etc/docker/daemon.json
fi

cat /etc/docker/daemon.json
sudo systemctl daemon-reload
sudo systemctl restart docker
sudo systemctl status docker
#                                                               #
#---------------------------------------------------------------#


if [ "$2" = 'LDAP' ]; then
    cd tests && sudo ./ldapprepare.sh && cd ..
fi

python --version
pip -V

#sudo apt-get update && sudo apt-get install -y --no-install-recommends libssl-dev && sudo apt-get autoremove -y && sudo rm -rf /var/lib/apt/lists/*
sudo wget https://bootstrap.pypa.io/get-pip.py && sudo python ./get-pip.py && sudo pip install --ignore-installed urllib3 chardet requests --upgrade
sudo make build_base_images -e BASEIMAGETAG=dev
sudo make install GOBUILDIMAGE=golang:1.17.9 COMPILETAG=compile_golangimage CLARITYIMAGE=goharbor/harbor-clarity-ui-builder:1.6.0 NOTARYFLAG=true CLAIRFLAG=true CHARTFLAG=true BUILDBIN=true PULL_BASE_FROM_DOCKERHUB=false
sleep 10
