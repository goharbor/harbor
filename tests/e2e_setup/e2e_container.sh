#!/bin/bash

HARBOR_SRC_FOLDER=$(realpath ../../)
echo ${HARBOR_SRC_FOLDER}

docker run -it --privileged -v /var/log/harbor:/var/log/harbor -v /etc/hosts:/etc/hosts -v ${HARBOR_SRC_FOLDER}:/drone -v ${HARBOR_SRC_FOLDER}/tests/harbor_ca.crt:/ca/ca.crt -v /dev/shm:/dev/shm -w /drone registry.goharbor.io/harbor-ci/goharbor/harbor-e2e-engine:latest-ui /bin/bash

