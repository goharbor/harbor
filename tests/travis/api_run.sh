#!/bin/bash

set -e

docker ps
# run db auth api cases
if [ "$1" = 'DB' ]; then
    pybot -v ip:$2 -v HARBOR_PASSWORD:Harbor12345 /home/travis/gopath/src/github.com/goharbor/harbor/tests/robot-cases/Group0-BAT/API_DB.robot
fi
# run ldap api cases
if [ "$1" = 'LDAP' ]; then
    pybot -v ip:$2 -v HARBOR_PASSWORD:Harbor12345 /home/travis/gopath/src/github.com/goharbor/harbor/tests/robot-cases/Group0-BAT/API_LDAP.robot
fi
