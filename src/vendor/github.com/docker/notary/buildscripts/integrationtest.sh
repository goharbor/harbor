#!/usr/bin/env bash

db="$1"
case ${db} in
  mysql*)
    db="mysql"
    ;;
  rethink*)
    db="rethink"
    ;;
  postgresql*)
    db="postgresql"
    ;;
  *)
    echo "Usage: $0 (mysql|rethink|postgresql)"
    exit 1
    ;;
esac

composeFile="development.${db}.yml"
project=integration

function cleanup {
    rm -f bin/notary
	docker-compose -p "${project}_${db}" -f ${composeFile} kill
	# if we're in CircleCI, we cannot remove any containers
	if [[ -z "${CIRCLECI}" ]]; then
		docker-compose -p "${project}_${db}" -f ${composeFile} down -v --remove-orphans
	fi
}

function cleanupAndExit {
    cleanup
    # Check for existence of SUCCESS
    ls test_output/SUCCESS
    exitCode=$?
    # Clean up test_output dir (if not in CircleCI) and exit
    if [[ -z "${CIRCLECI}" ]]; then
        rm -rf test_output
    fi
    exit $exitCode
}

if [[ -z "${CIRCLECI}" ]]; then
	BUILDOPTS="--force-rm"
fi

set -e
set -x

cleanup

docker-compose -p "${project}_${db}" -f ${composeFile} config
docker-compose -p "${project}_${db}" -f ${composeFile} build ${BUILDOPTS} --pull | tee

trap cleanupAndExit SIGINT SIGTERM EXIT

docker-compose -p "${project}_${db}" -f ${composeFile} up --abort-on-container-exit
