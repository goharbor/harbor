#!/bin/bash

sudo gsutil version -l

harbor_logs_bucket="harbor-ci-logs"

# GC credentials
botofile="/home/travis/.boto"
echo "[Credentials]" >> $botofile
echo "gs_access_key_id = $gs_access_key_id" >> $botofile
echo "gs_secret_access_key = $gs_secret_access_key" >> $botofile
echo "[GSUtil]" >> $botofile
echo "content_language = en" >> $botofile
echo "default_project_id = $default_project_id" >> $botofile

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
outfile="integration_logs_$TRAVIS_BUILD_NUMBER_$TRAVIS_COMMIT.tar.gz"
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
