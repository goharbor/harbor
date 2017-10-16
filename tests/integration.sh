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

## --------------------------------------------- Init Env -------------------------------------------------
dpkg -l > package.list
# Start Xvfb for Chrome headlesss
Xvfb -ac :99 -screen 0 1280x1024x16 & export DISPLAY=:99

export DRONE_SERVER=$DRONE_SERVER
export DRONE_TOKEN=$DRONE_TOKEN
buildinfo=$(drone build info vmware/harbor $DRONE_BUILD_NUMBER)
echo $buildinfo
upload_build=false
nightly_run=false
upload_latest_build=false
latest_build_file='latest.build'

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
container_ip=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
echo $container_ip

## --------------------------------------------- Run Test Case ---------------------------------------------
if [ $DRONE_REPO != "vmware/harbor" ]; then
  echo "Only run tests again Harbor Repo."
  exit 1
fi

if [[ $DRONE_BRANCH == "master" || $DRONE_BRANCH == *"refs/tags"* || $DRONE_BRANCH == "release-"* ]] && [[ $DRONE_BUILD_EVENT == "push" || $DRONE_BUILD_EVENT == "tag" ]]; then
    ## -------------- Package installer with clean code -----------------
    echo "Package Harbor build."
    pybot --removekeywords TAG:secret --include Bundle tests/robot-cases/Group0-Distro-Harbor
    echo "Running full CI for $DRONE_BUILD_EVENT on $DRONE_BRANCH"
    pybot -v ip:$container_ip --removekeywords TAG:secret --include BAT tests/robot-cases/Group0-BAT
    upload_latest_build=true
elif (echo $buildinfo | grep -q "\[Specific CI="); then
    buildtype=$(echo $buildinfo | grep "\[Specific CI=")
    testsuite=$(echo $buildtype | awk -v FS="(=|])" '{print $2}')
    pybot -v ip:$container_ip --removekeywords TAG:secret --suite $testsuite --suite Regression tests/robot-cases
elif (echo $buildinfo | grep -q "\[Full CI\]"); then
    upload_build=true
    pybot -v ip:$container_ip --removekeywords TAG:secret --exclude skip tests/robot-cases
elif (echo $buildinfo | grep -q "\[Skip CI\]"); then
    echo "Skip CI."
else
	# default mode is BAT.
    pybot -v ip:$container_ip --removekeywords TAG:secret --include BAT tests/robot-cases/Group0-BAT
fi

rc="$?"
echo $rc

timestamp=$(date +%s)
outfile="integration_logs_"$DRONE_BUILD_NUMBER"_"$DRONE_COMMIT".tar.gz"
tar -zcvf $outfile output.xml log.html *.png package.list *container-logs.zip *.log /var/log/harbor/* /data/config/* /data/job_logs/*
if [ -f "$outfile" ]; then
  gsutil cp $outfile gs://harbor-ci-logs
  echo "----------------------------------------------"
  echo "Download test logs:"
  echo "https://storage.googleapis.com/harbor-ci-logs/$outfile"
  echo "----------------------------------------------"
  gsutil -D setacl public-read gs://harbor-ci-logs/$outfile &> /dev/null
else
  echo "No log output file to upload"
fi

## --------------------------------------------- Upload Harbor Latest Build File ---------------------------------------
if [ $upload_latest_build == true ] && [ $rc -eq 0 ]; then
  harbor_build_bundle=$(basename harbor-offline-installer-*.tgz)
  if [[ $DRONE_BRANCH == "master" || $DRONE_BRANCH == *"refs/tags"* || $DRONE_BRANCH == "release-"* ]] && [[ $DRONE_BUILD_EVENT == "push" || $DRONE_BUILD_EVENT == "tag" ]]; then
      echo 'https://storage.googleapis.com/harbor-builds/$harbor_build_bundle' > $latest_build_file
      gsutil cp $latest_build_file gs://harbor-builds
      gsutil -D setacl public-read gs://harbor-builds/$latest_build_file &> /dev/null
  fi
  if [[ $DRONE_BRANCH == *"refs/tags"* || $DRONE_BRANCH == "release-"* ]] && [[ $DRONE_BUILD_EVENT == "push" || $DRONE_BUILD_EVENT == "tag" ]]; then
      echo 'https://storage.googleapis.com/harbor-releases/$harbor_build_bundle' > $latest_build_file
      gsutil cp $latest_build_file gs://harbor-releases
      gsutil -D setacl public-read gs://harbor-releases/$latest_build_file &> /dev/null
  fi    
fi

## --------------------------------------------- Sendout Email ---------------------------------------------
if [ $nightly_run == true ]; then
    echo "Sendout Nightly Run Email."
    if [ $rc -eq 0 ]; then
        result=Pass
    else
        result=Fail
    fi
    python tests/nightly/sendreport.py --repo $DRONE_REPO --branch $DRONE_BRANCH --commit $DRONE_COMMIT --result $result --log $outfile --mailpwd $MAIL_PWD
    echo "Sendout Nightly Run Email success."
fi

## --------------------------------------------- Tear Down -------------------------------------------------
if [ -f "$keyfile" ]; then
  rm -f $keyfile
fi

exit $rc