#!/bin/bash
set -x

set +e
sudo rm -fr /data/*
sudo mkdir -p /data
DIR="$(cd "$(dirname "$0")" && pwd)"

set -e
if [ -z "$1" ]; then echo no ip specified; exit 1;fi
# prepare cert ...
sudo ./tests/generateCerts.sh $1

python --version
pip -V
cat /etc/issue
cat /proc/version
sudo -H pip install --ignore-installed urllib3 chardet requests --upgrade
python --version

#---------------Set DNS for docker v20-----------------------#
# In docker v20, it fixed an issue named  "Wrong resolv.conf
# used on Ubuntu 19", this fix caused DNS solve problem
# in container. So the current work round is read DNS server
# from system and set the value in /etc/docker/daemon.json.

ip addr
docker_config_file="/etc/docker/daemon.json"
dns_ip_string=$(netplan ip leases eth0 | grep -i dns | awk -F = '{print $2}' | tr " " "\n" | sed 's/,/","/g')
dns=[\"${dns_ip_string}\"]
echo dns=${dns}

cat $docker_config_file
if [ -f $docker_config_file ];then
    if [ $(cat /etc/docker/daemon.json |grep \"dns\" |wc -l) -eq 0 ];then
        sudo sed "s/}/,\n   \"dns\": $dns\n}/" -i $docker_config_file
    fi
else
    echo "{\"dns\": $dns}" > $docker_config_file
fi
cat $docker_config_file
sudo systemctl stop docker
sudo systemctl start docker
sleep 2
#------------------------------------------------------------#

sudo ./tests/hostcfg.sh

if [ "$2" = 'LDAP' ]; then
    cd tests && sudo ./ldapprepare.sh && cd ..
fi

if [ $GITHUB_TOKEN ];
then
    sed "s/# github_token: xxx/github_token: $GITHUB_TOKEN/" -i make/harbor.yml
fi

sudo make compile build prepare COMPILETAG=compile_golangimage GOBUILDTAGS="include_oss include_gcs" NOTARYFLAG=true TRIVYFLAG=true CHARTFLAG=true GEN_TLS=true PULL_BASE_FROM_DOCKERHUB=false

# set the debugging env
echo "GC_TIME_WINDOW_HOURS=0" | sudo tee -a ./make/common/config/core/env
sudo make start

# waiting 5 minutes to start
for((i=1;i<=30;i++)); do
  echo $i waiting 10 seconds...
  sleep 10
  curl -k -L -f 127.0.0.1/api/v2.0/systeminfo && break
  docker ps
done
