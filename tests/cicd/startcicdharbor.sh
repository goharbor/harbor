#!/usr/bin/env bash

if [ -z "$2" ];then echo "$0 <ip> <buildnum> [http_port] [https_port]";exit 1;fi
IP=$1
BUILDNUM=$2
HTTP_PORT=${3:-80}
HTTPS_PORT=${4:-443}

TAG=build.$BUILDNUM
NAMESPACE="cicd.harbor.bitsf.xin/harbor-dev"
data_path=$(pwd)/$TAG/data
mkdir -p $data_path
config_dir=$(pwd)/$TAG/common/config
mkdir -p $config_dir
mkdir -p $data_path/logs
compose_file=$(pwd)/$TAG/docker-compose.yml
touch $compose_file
secret_dir=$data_path/secret
mkdir -p $secret_dir
cert_path=$data_path/cert
mkdir -p $cert_path

docker pull $NAMESPACE/registry-photon:v2.7.1-patch-2819
docker tag $NAMESPACE/registry-photon:v2.7.1-patch-2819 $NAMESPACE/registry-photon:v2.7.1-patch-2819-$TAG
for name in prepare harbor-registryctl nginx-photon harbor-portal harbor-jobservice harbor-core harbor-db redis-photon harbor-log; do
  docker pull $NAMESPACE/$name:$TAG
done

curl https://raw.githubusercontent.com/goharbor/harbor/master/tests/harbor_ca.key -o $cert_path/harbor_ca.key
curl https://raw.githubusercontent.com/goharbor/harbor/master/tests/harbor_ca.crt -o $cert_path/harbor_ca.crt
openssl req \
    -newkey rsa:4096 -nodes -sha256 -keyout $cert_path/$IP.key \
    -out $cert_path/$IP.csr -subj "/C=CN/ST=PEK/L=Bei Jing/O=VMware/CN=HarborManager"
echo subjectAltName = IP:$IP > $cert_path/extfile.cnf
openssl x509 -req -days 365 -sha256 -in $cert_path/$IP.csr -CA $cert_path/harbor_ca.crt \
	-CAkey $cert_path/harbor_ca.key -CAcreateserial -CAserial $cert_path/$IP.srl -extfile $cert_path/extfile.cnf -out $cert_path/$IP.crt

docker run --rm -v $(pwd)/fixcicdharbor.py:/usr/src/app/fixcicdharbor.py \
                    -v $data_path:/data:z \
                    -v $compose_file:/compose_location/docker-compose.yml:z \
                    -v $config_dir:/config:z \
                    -v $secret_dir:/secret:z \
                    -v $cert_path/$IP.key:/hostfs/cert/server.key:z \
                    -v $cert_path/$IP.crt:/hostfs/cert/server.crt:z \
                    -e IP=$IP -e HTTP_PORT=$HTTP_PORT -e HTTPS_PORT=$HTTPS_PORT \
                    -e data_volume=$data_path \
                    -e TAG=$TAG -e NAMESPACE=$NAMESPACE \
                    --entrypoint ./fixcicdharbor.py \
                    $NAMESPACE/prepare:$TAG \
                    || exit 1

sudo chmod -R +r $TAG
sudo chmod -R 700 $data_path/database

cd $TAG
docker-compose down
docker-compose up -d

echo enjoy you harbor at http://$IP:$HTTP_PORT
