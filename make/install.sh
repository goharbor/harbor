#!/bin/bash

set -e

DIR="$(cd "$(dirname "$0")" && pwd)"
source $DIR/common.sh

set +o noglob

usage=$'Please set hostname and other necessary attributes in harbor.yml first. DO NOT use localhost or 127.0.0.1 for hostname, because Harbor needs to be accessed by external clients.
Please set --with-trivy if needs enable Trivy in Harbor.
Please do NOT set --with-chartmuseum, as chartmusuem has been deprecated and removed.'
item=0

# clair is deprecated
with_clair=$false
# trivy is not enabled by default
with_trivy=$false
# assume no systemd for now
have_systemd=$false

# flag to using docker compose v1 or v2, default would using v1 docker-compose
DOCKER_COMPOSE=docker-compose

while [ $# -gt 0 ]; do
        case $1 in
            --help)
            note "$usage"
            exit 0;;
            --with-trivy)
            with_trivy=true;;
            *)
            note "$usage"
            exit 1;;
        esac
        shift || true
done

workdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $workdir

h2 "[Step $item]: checking if docker is installed ..."; let item+=1
check_docker

h2 "[Step $item]: checking docker-compose is installed ..."; let item+=1
check_dockercompose

if [ -f harbor*.tar.gz ]
then
    h2 "[Step $item]: loading Harbor images ..."; let item+=1
    docker load -i ./harbor*.tar.gz
fi
echo ""

h2 "[Step $item]: preparing environment ...";  let item+=1
if [ -n "$host" ]
then
    sed "s/^hostname: .*/hostname: $host/g" -i ./harbor.yml
fi

h2 "[Step $item]: preparing harbor configs ...";  let item+=1
prepare_para=
if [ $with_trivy ]
then
    prepare_para="${prepare_para} --with-trivy"
fi

./prepare $prepare_para
echo ""

if [ -n "$DOCKER_COMPOSE ps -q"  ]
    then
        note "stopping existing Harbor instance ..." 
        $DOCKER_COMPOSE down -v
fi
echo ""

if [ -d /etc/systemd ]
then
    have_systemd=true
    h2 "[Step $item]: installing Harbor systemd service ..."; let item+=1

    cat >/etc/systemd/system/harbor.service <<EOF
[Unit]
Description=Harbor Cloud Native Registry
Documentation=https://goharbor.io
After=docker.service
Requires=docker.service

[Service]
Type=simple
Restart=on-failure
RestartSec=5
ExecStart=${DOCKER_COMPOSE} -f ${workdir}/docker-compose.yml up
ExecStop=${DOCKER_COMPOSE} -f ${workdir}/docker-compose.yml down -v
ExecStopPost=${DOCKER_COMPOSE} -f ${workdir}/docker-compose.yml rm -f

[Install]
WantedBy=multi-user.target
EOF

    note "Reloading systemd unit files ..."
    systemctl daemon-reload

    note "Setting Harbor to start on boot ..."
    systemctl enable harbor
fi

h2 "[Step $item]: starting Harbor ..."; let item+=1

if [ $have_systemd ]
then
    systemctl start harbor
else
    $DOCKER_COMPOSE up -d
fi

success $"----Harbor has been installed and started successfully.----"
