#!/bin/bash
set -x

set -e

export POSTGRESQL_HOST=$1
export REGISTRY_URL=$1:5000
export CHROME_BIN=chromium-browser
#export DISPLAY=:99.0
#sh -e /etc/init.d/xvfb start

sudo docker-compose -f ./make/docker-compose.test.yml up -d
sleep 10
./tests/pushimage.sh
docker ps

DIR="$(cd "$(dirname "$0")" && pwd)"
go test -race -i ./src/core ./src/jobservice
sudo -E env "PATH=$PATH" "POSTGRES_MIGRATION_SCRIPTS_PATH=$DIR/../../make/migrations/postgresql/" ./tests/coverage4gotest.sh
find .|grep profile.cov