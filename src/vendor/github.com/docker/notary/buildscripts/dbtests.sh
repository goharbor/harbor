#!/usr/bin/env bash

db="$1"

case ${db} in
  mysql*)
    db="mysql"
    dbContainerOpts="--name mysql_tests mysql mysqld --innodb_file_per_table"
    DBURL="server@tcp(mysql_tests:3306)/notaryserver?parseTime=True"
    ;;
  rethink*)
    db="rethink"
    dbContainerOpts="--name rethinkdb_tests rdb-01 --bind all --driver-tls-key /tls/key.pem --driver-tls-cert /tls/cert.pem"
    DBURL="rethinkdb_tests"
    ;;
  postgresql*)
    db="postgresql"
    dbContainerOpts="--name postgresql_tests postgresql"
    DBURL="postgres://server@postgresql_tests:5432/notaryserver?sslmode=disable"
    ;;
  *)
    echo "Usage: $0 (mysql|rethink)"
    exit 1
    ;;
esac

composeFile="development.${db}.yml"
project=dbtests

function cleanup {
    rm -f bin/notary
    docker-compose -p "${project}_${db}" -f "${composeFile}" kill
    # if we're in CircleCI, we cannot remove any containers
    if [[ -z "${CIRCLECI}" ]]; then
        docker-compose -p "${project}_${db}" -f "${composeFile}" down -v --remove-orphans
    fi
}

clientCmd="make TESTOPTS='-p 1' test"
if [[ -z "${CIRCLECI}" ]]; then
    BUILDOPTS="--force-rm"
else
    clientCmd="make ci && codecov"
fi

set -e
set -x

cleanup

docker-compose -p "${project}_${db}" -f ${composeFile} build ${BUILDOPTS} client

trap cleanup SIGINT SIGTERM EXIT

# run the unit tests that require a DB

docker-compose -p "${project}_${db}" -f "${composeFile}" run --no-deps -d ${dbContainerOpts}
docker-compose -p "${project}_${db}" -f "${composeFile}" run --no-deps \
    -e NOTARY_BUILDTAGS="${db}db" -e DBURL="${DBURL}" \
    -e PKGS="github.com/docker/notary/server/storage github.com/docker/notary/signer/keydbstore" \
    client bash -c "${clientCmd}"
