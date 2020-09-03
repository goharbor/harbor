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
sudo mkdir -p /etc/docker/certs.d/$1 && sudo cp ./tests/harbor_ca.crt /etc/docker/certs.d/$1/ && rm -rf ~/.docker/ &&  mkdir -p ~/.docker/tls/$1:4443/ && sudo cp ./tests/harbor_ca.crt ~/.docker/tls/$1:4443/

sudo ./tests/hostcfg.sh

if [ "$2" = 'LDAP' ]; then
    cd tests && sudo ./ldapprepare.sh && cd ..
fi

# prepare a chart file for API_DB test...
sudo curl -o $DIR/../../tests/apitests/python/mariadb-4.3.1.tgz https://storage.googleapis.com/harbor-builds/bin/charts/mariadb-4.3.1.tgz

sudo wget https://bootstrap.pypa.io/get-pip.py && sudo python ./get-pip.py && sudo pip install --ignore-installed urllib3 chardet requests && sudo pip install robotframework==3.2.1 robotframework-httplibrary requests --upgrade
sudo make swagger_client
#TODO: Swagger python package used to installed into dist-packages, but it's changed into site-packages all in a sudden, we havn't found the root cause.
#      so current workround is to copy swagger packages from site-packages to dist-packages.
package_dir=/usr/lib/python3.7/site-packages
if [ $(find $package_dir -type f| wc -l) -gt 0 ];then
    sudo cp -r ${package_dir}/* /usr/local/lib/python3.7/dist-packages
fi

if [ $GITHUB_TOKEN ];
then
    sed "s/# github_token: xxx/github_token: $GITHUB_TOKEN/" -i make/harbor.yml
fi

sudo make build_base_docker compile build prepare COMPILETAG=compile_golangimage GOBUILDTAGS="include_oss include_gcs" NOTARYFLAG=true CLAIRFLAG=true TRIVYFLAG=true CHARTFLAG=true GEN_TLS=true

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
