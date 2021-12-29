#!/bin/bash

set -e

DIR="$(cd "$(dirname "$0")" && pwd)"
source $DIR/common.sh

set +o noglob

usage=$'Please set hostname and other necessary attributes in harbor.yml first. DO NOT use localhost or 127.0.0.1 for hostname, because Harbor needs to be accessed by external clients.
Please set --with-notary if needs enable Notary in Harbor, and set ui_url_protocol/ssl_cert/ssl_cert_key in harbor.yml bacause notary must run under https. 
Please set --with-trivy if needs enable Trivy in Harbor
Please set --with-chartmuseum if needs enable Chartmuseum in Harbor
Please set --offline-installer-file /path/to/harbor-offline-installer-vX.X.X.tgz if not in the same directory as install.sh.'
item=0

# notary is not enabled by default
with_notary=$false
# clair is deprecated
with_clair=$false
# trivy is not enabled by default
with_trivy=$false
# chartmuseum is not enabled by default
with_chartmuseum=$false

while [ $# -gt 0 ]; do
        case $1 in
            --help)
            note "$usage"
            exit 0;;
            --with-notary)
            with_notary=true;;
            --with-clair)
            with_clair=true;;
            --with-trivy)
            with_trivy=true;;
            --with-chartmuseum)
            with_chartmuseum=true;;
            --offline-installer-file)
            INSTALLERFILE="$2" ; shift;;
            *)
            note "$usage"
            exit 1;;
        esac
        shift || true
done

if [ $with_clair ]
then
    error "Clair is deprecated please remove it from installation arguments !!!"
    exit 1
fi

workdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $workdir

h2 "[Step $item]: checking if docker is installed ..."; let item+=1
check_docker

h2 "[Step $item]: checking docker-compose is installed ..."; let item+=1
check_dockercompose

if [[ $INSTALLERFILE == "" ]]; then
    _INSTALLERFILE="./harbor*.tar.gz"
else
    _INSTALLERFILE=$INSTALLERFILE
fi

if [ -f $_INSTALLERFILE ]; then
    h2 "[Step $item]: loading Harbor images ..."; let item+=1
    docker load -i $_INSTALLERFILE
fi

echo ""

h2 "[Step $item]: preparing environment ...";  let item+=1
if [ -n "$host" ]
then
    sed "s/^hostname: .*/hostname: $host/g" -i ./harbor.yml
fi

h2 "[Step $item]: preparing harbor configs ...";  let item+=1
prepare_para=
if [ $with_notary ] 
then
    prepare_para="${prepare_para} --with-notary"
fi
if [ $with_trivy ]
then
    prepare_para="${prepare_para} --with-trivy"
fi
if [ $with_chartmuseum ]
then
    prepare_para="${prepare_para} --with-chartmuseum"
fi

./prepare $prepare_para
echo ""

if [ -n "$(docker-compose ps -q)"  ]
then
    note "stopping existing Harbor instance ..." 
    docker-compose down -v
fi
echo ""

h2 "[Step $item]: starting Harbor ..."
docker-compose up -d

success $"----Harbor has been installed and started successfully.----"
