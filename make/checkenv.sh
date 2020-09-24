#!/bin/bash

set -o noglob
set -e

usage=$'Checking environment for harbor build and install. Include golang, docker and docker-compose.'

while [ $# -gt 0 ]; do
        case $1 in
            --help)
            note "$usage"
            exit 0;;
            *)
            note "$usage"
            exit 1;;
        esac
        shift || true
done

DIR="$(cd "$(dirname "$0")" && pwd)"
source $DIR/common.sh

check_golang
check_docker
check_dockercompose
