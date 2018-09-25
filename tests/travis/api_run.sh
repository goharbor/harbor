#!/bin/bash

source gskey.sh

sudo gsutil version -l

harbor_logs_bucket="harbor-ci-logs"
# GC credentials
keyfile="/home/travis/harbor-ci-logs.key"
botofile="/home/travis/.boto"
sudo echo -en $GS_PRIVATE_KEY > $keyfile
sudo chmod 400 $keyfile
sudo echo "[Credentials]" >> $botofile
sudo echo "gs_service_key_file = $keyfile" >> $botofile
sudo echo "gs_service_client_id = $GS_CLIENT_EMAIL" >> $botofile
sudo echo "[GSUtil]" >> $botofile
sudo echo "content_language = en" >> $botofile
sudo echo "default_project_id = $GS_PROJECT_ID" >> $botofile

# GS util
function uploader {
    sudo gsutil cp $1 gs://$2/$1
    sudo gsutil -D setacl public-read gs://$2/$1 &> /dev/null
}

set +e

docker ps
# run db auth api cases
if [ "$1" = 'DB' ]; then
    pybot -v ip:$2 -v HARBOR_PASSWORD:Harbor12345 /home/travis/gopath/src/github.com/goharbor/harbor/tests/robot-cases/Group0-BAT/API_DB.robot
fi
# run ldap api cases
if [ "$1" = 'LDAP' ]; then
    pybot -v ip:$2 -v HARBOR_PASSWORD:Harbor12345 /home/travis/gopath/src/github.com/goharbor/harbor/tests/robot-cases/Group0-BAT/API_LDAP.robot
fi

## --------------------------------------------- Upload Harbor CI Logs -------------------------------------------
timestamp=$(date +%s)
outfile="integration_logs_"$TRAVIS_BUILD_NUMBER"_"$TRAVIS_COMMIT".tar.gz"
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
