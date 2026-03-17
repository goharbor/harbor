#!/bin/bash

set -e

DIR="$(cd "$(dirname "$0")" && pwd)"
source $DIR/common.sh

set +o noglob

usage=$'Please set hostname and other necessary attributes in harbor.yml first. DO NOT use localhost or 127.0.0.1 for hostname, because Harbor needs to be accessed by external clients.
Please set --with-trivy if needs enable Trivy in Harbor.
Please set --with-podman if you want to force prepare to treat the Docker CLI as Podman.
Please do NOT set --with-chartmuseum, as chartmusuem has been deprecated and removed.
Please do NOT set --with-notary, as notary has been deprecated and removed.'
item=0

# clair is deprecated
with_clair=$false
# trivy is not enabled by default
with_trivy=$false
# podman option not enabled by default
with_podman=$false

# flag to using docker compose v1 or v2, default would using v1 docker-compose
DOCKER_COMPOSE=docker-compose

while [ $# -gt 0 ]; do
        case $1 in
            --help)
            note "$usage"
            exit 0;;
            --with-trivy)
            with_trivy=true;;
            --with-podman)
            with_podman=true;;
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

if [ $with_podman ]
then
    prepare_para="${prepare_para} --with-podman"
fi

./prepare $prepare_para

echo ""

# Check if any containers started by compose are present. Capture the compose
# output and filter to only valid container IDs so vendor warnings (for
# example "Emulate Docker CLI using podman...") don't make the check true.
compose_ps_output=$($DOCKER_COMPOSE ps -q 2>/dev/null | grep -E '^[0-9a-fA-F]+$' || true)
if [ -n "$compose_ps_output" ]
then
    note "stopping existing Harbor instance ..."
    $DOCKER_COMPOSE down -v
fi
echo ""

h2 "[Step $item]: starting Harbor ..."
$DOCKER_COMPOSE up -d

success $"----Harbor has been installed and started successfully.----"
