#!/bin/bash
set -x

#source gskey.sh

sudo gsutil version -l

harbor_logs_bucket="harbor-ci-logs"

DIR="$(cd "$(dirname "$0")" && pwd)"
E2E_IMAGE="goharbor/harbor-e2e-engine:4.2.1-api"

# GS util
function uploader {
   sudo gsutil cp $1 gs://$2/$1
   sudo gsutil acl ch -u AllUsers:R gs://$2/$1
}

set +e

docker ps
# run db auth api cases
if [ "$1" = 'DB' ]; then
    docker run -i --privileged -v $DIR/../../:/drone -v $DIR/../:/ca -w /drone $E2E_IMAGE robot --exclude proxy_cache -v DOCKER_USER:${DOCKER_USER} -v DOCKER_PWD:${DOCKER_PWD} -v ip:$2  -v ip1: -v http_get_ca:false -v HARBOR_PASSWORD:Harbor12345 /drone/tests/robot-cases/Group1-Nightly/Setup.robot /drone/tests/robot-cases/Group0-BAT/API_DB.robot
elif [ "$1" = 'PROXY_CACHE' ]; then
    docker run -i --privileged -v $DIR/../../:/drone -v $DIR/../:/ca -w /drone $E2E_IMAGE robot --include setup  --include proxy_cache -v DOCKER_USER:${DOCKER_USER} -v DOCKER_PWD:${DOCKER_PWD} -v ip:$2  -v ip1: -v http_get_ca:false -v HARBOR_PASSWORD:Harbor12345 /drone/tests/robot-cases/Group1-Nightly/Setup.robot /drone/tests/robot-cases/Group0-BAT/API_DB.robot
elif [ "$1" = 'LDAP' ]; then
    # run ldap api cases
    python $DIR/../../tests/configharbor.py -H $IP -u $HARBOR_ADMIN -p $HARBOR_ADMIN_PASSWD -c auth_mode=ldap_auth \
                                  ldap_url=ldap://$IP \
                                  ldap_search_dn=cn=admin,dc=example,dc=com \
                                  ldap_search_password=admin \
                                  ldap_base_dn=dc=example,dc=com \
                                  ldap_uid=cn
    docker run -i --privileged -v $DIR/../../:/drone -v $DIR/../:/ca -w /drone $E2E_IMAGE robot -v DOCKER_USER:${DOCKER_USER} -v DOCKER_PWD:${DOCKER_PWD} -v ip:$2  -v ip1: -v http_get_ca:false -v HARBOR_PASSWORD:Harbor12345 /drone/tests/robot-cases/Group1-Nightly/Setup.robot /drone/tests/robot-cases/Group0-BAT/API_LDAP.robot
else
    rc=999
fi
rc=$?
## --------------------------------------------- Upload Harbor CI Logs -------------------------------------------
timestamp=$(date +%s)
GIT_COMMIT=$(git rev-parse --short "$GITHUB_SHA")
outfile="integration_logs_$timestamp$GIT_COMMIT.tar.gz"
sudo tar -zcvf $outfile output.xml log.html /var/log/harbor/*
if [ -f "$outfile" ]; then
   uploader $outfile $harbor_logs_bucket
   echo "----------------------------------------------"
   echo "Download test logs:"
   echo "https://storage.googleapis.com/harbor-ci-logs/$outfile"
   echo "----------------------------------------------"
else
   echo "No log output file to upload"
fi

exit $rc
