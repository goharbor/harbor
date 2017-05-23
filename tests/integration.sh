#!/bin/bash
# Copyright 2016 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -x
gsutil version -l
set +x

dpkg -l > package.list

buildinfo=$(drone build info vmware/harbor $DRONE_BUILD_NUMBER)

Xvfb -ac :99 -screen 0 1280x1024x16 & export DISPLAY=:99

if [ $DRONE_BRANCH = "master" ] && [ $DRONE_REPO = "vmware/harbor" ]; then
    pybot --removekeywords TAG:secret --include BAT tests/robot-cases/Group0-BAT
elif grep -q "\[full ci\]" <(drone build info vmware/harbor $DRONE_BUILD_NUMBER); then
    pybot --removekeywords TAG:secret --exclude skip tests/robot-cases
elif (echo $buildinfo | grep -q "\[specific ci="); then
    buildtype=$(echo $buildinfo | grep "\[specific ci=")
    testsuite=$(echo $buildtype | awk -v FS="(=|])" '{print $2}')
    pybot --removekeywords TAG:secret --suite $testsuite --suite Regression tests/robot-cases
else
    pybot --removekeywords TAG:secret --include BAT tests/robot-cases/Group0-BAT
fi

rc="$?"

timestamp=$(date +%s)
outfile="integration_logs_"$DRONE_BUILD_NUMBER"_"$DRONE_COMMIT".zip"

zip -9 $outfile output.xml log.html package.list *container-logs.zip *.log

# GC credentials
keyfile="/root/harbor-ci-logs.key"
botofile="/root/.boto"
echo -en $GS_PRIVATE_KEY > $keyfile
chmod 400 $keyfile
echo "[Credentials]" >> $botofile
echo "gs_service_key_file = $keyfile" >> $botofile
echo "gs_service_client_id = $GS_CLIENT_EMAIL" >> $botofile
echo "[GSUtil]" >> $botofile
echo "content_language = en" >> $botofile
echo "default_project_id = $GS_PROJECT_ID" >> $botofile

if [ -f "$outfile" ]; then
  gsutil cp $outfile gs://harbor-ci-logs
  echo "----------------------------------------------"
  echo "Download test logs:"
  echo "https://console.cloud.google.com/m/cloudstorage/b/harbor-ci-logs/o/$outfile?authuser=1"
  echo "----------------------------------------------"
  gsutil -D setacl public-read gs://harbor-ci-logs/$outfile >/dev/null
else
  echo "No log output file to upload"
fi

if [ -f "$keyfile" ]; then
  rm -f $keyfile
fi

exit $rc
