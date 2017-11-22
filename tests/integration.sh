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

harbor_logs_bucket="harbor-ci-logs"
harbor_builds_bucket="harbor-builds"
harbor_releases_bucket="harbor-releases"
harbor_ci_pipeline_store_bucket="harbor-ci-pipeline-store/latest"

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

# GS util
function uploader {
  gsutil cp $1 gs://$2
  gsutil -D setacl public-read gs://$2/$1 &> /dev/null
}

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
    upload_latest_build=true
    pybot -v ip:$container_ip --removekeywords TAG:secret --include BAT tests/robot-cases/Group0-BAT
elif (echo $buildinfo | grep -q "\[Specific CI="); then
    buildtype=$(echo $buildinfo | grep "\[Specific CI=")
    testsuite=$(echo $buildtype | awk -v FS="(=|])" '{print $2}')
    pybot -v ip:$container_ip --removekeywords TAG:secret --suite $testsuite --suite Regression tests/robot-cases
elif (echo $buildinfo | grep -q "\[Full CI\]"); then
    pybot -v ip:$container_ip --removekeywords TAG:secret --exclude skip tests/robot-cases
elif (echo $buildinfo | grep -q "\[Skip CI\]"); then
    echo "Skip CI."
elif (echo $buildinfo | grep -q "\[Upload Build\]"); then
    upload_latest_build=true
    upload_build=true
    echo "Package Harbor build."
    pybot --removekeywords TAG:secret --include Bundle tests/robot-cases/Group0-Distro-Harbor
    echo "Running full CI for $DRONE_BUILD_EVENT on $DRONE_BRANCH"
    pybot -v ip:$container_ip --removekeywords TAG:secret --include BAT tests/robot-cases/Group0-BAT
else
    # default mode is BAT.
    pybot -v ip:$container_ip --removekeywords TAG:secret --include BAT tests/robot-cases/Group0-BAT
fi

rc="$?"
echo $rc

## --------------------------------------------- Upload Harbor CI Logs -------------------------------------------
timestamp=$(date +%s)
outfile="integration_logs_"$DRONE_BUILD_NUMBER"_"$DRONE_COMMIT".tar.gz"
tar -zcvf $outfile output.xml log.html *.png package.list *container-logs.zip *.log /var/log/harbor/* /data/config/* /data/job_logs/*
if [ -f "$outfile" ]; then
  uploader $outfile $harbor_logs_bucket
  echo "----------------------------------------------"
  echo "Download test logs:"
  echo "https://storage.googleapis.com/harbor-ci-logs/$outfile"
  echo "----------------------------------------------"
else
  echo "No log output file to upload"
fi

## --------------------------------------------- Upload Harbor Bundle File ---------------------------------------
if [ $upload_build == true ] && [ $rc -eq 0 ]; then
  harbor_build_bundle=$(basename harbor-offline-installer-*.tgz)
  uploader $harbor_build_bundle $harbor_builds_bucket

  if [ $DRONE_BRANCH == "master" ]; then
    cp $harbor_build_bundle harbor-offline-installer-latest-master.tgz
    uploader harbor-offline-installer-latest-master.tgz $harbor_ci_pipeline_store_bucket
  fi 
  if [[ $DRONE_BRANCH == *"refs/tags"* || $DRONE_BRANCH == "release-"* ]]; then
    cp $harbor_build_bundle harbor-offline-installer-latest-release.tgz
	uploader harbor-offline-installer-latest-release.tgz $harbor_ci_pipeline_store_bucket
  fi 
fi

## --------------------------------------------- Upload Harbor Latest Build File ---------------------------------
if [ $upload_latest_build == true ] && [ $rc -eq 0 ]; then
  echo "update latest build file."
  harbor_build_bundle=$(basename harbor-offline-installer-*.tgz)
  echo $harbor_build_bundle 
  if [[ $DRONE_BRANCH == "master" ]] && [[ $DRONE_BUILD_EVENT == "push" || $DRONE_BUILD_EVENT == "tag" ]]; then
      echo 'https://storage.googleapis.com/harbor-builds/'$harbor_build_bundle > $latest_build_file
      uploader $latest_build_file $harbor_builds_bucket	
  fi
  if [[ $DRONE_BRANCH == *"refs/tags"* || $DRONE_BRANCH == "release-"* ]] && [[ $DRONE_BUILD_EVENT == "push" || $DRONE_BUILD_EVENT == "tag" ]]; then
      echo 'https://storage.googleapis.com/harbor-releases/'$harbor_build_bundle > $latest_build_file
	  uploader $latest_build_file $harbor_releases_bucket
  fi    
fi

## --------------------------------------------- Tear Down -------------------------------------------------------
if [ -f "$keyfile" ]; then
  rm -f $keyfile
fi

exit $rc