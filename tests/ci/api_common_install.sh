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

echo "Generate swagger client"
wget https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/4.3.1/openapi-generator-cli-4.3.1.jar -O openapi-generator-cli.jar
rm -rf harborclient
mkdir  -p harborclient/harbor_client
mkdir  -p harborclient/harbor_swagger_client
mkdir  -p harborclient/harbor_v2_swagger_client
java -jar openapi-generator-cli.jar generate -i api/swagger.yaml -g python -o harborclient/harbor_client --package-name client
java -jar openapi-generator-cli.jar generate -i api/v2.0/legacy_swagger.yaml -g python -o harborclient/harbor_swagger_client --package-name swagger_client
java -jar openapi-generator-cli.jar generate -i api/v2.0/swagger.yaml -g python -o harborclient/harbor_v2_swagger_client --package-name v2_swagger_client
python ./harborclient/harbor_client/setup.py install --install-lib=$PYTHONPATH
python ./harborclient/harbor_swagger_client/setup.py install --install-lib=$PYTHONPATH
python ./harborclient/harbor_v2_swagger_client/setup.py install --install-lib=$PYTHONPATH
pip install docker -q
pip freeze
echo $PYTHONPATH
python -c 'import sys; print(sys.path)'
python -c 'import client;print(client.__file__)'
python -c 'import swagger_client;print(swagger_client.__file__)'
python -c 'import v2_swagger_client;print(v2_swagger_client.__file__)'
if [ -f $package_files ];then
    sudo cp -r $package_files /usr/local/lib/python3.7/dist-packages
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
