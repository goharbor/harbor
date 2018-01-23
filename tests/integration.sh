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
git_commit=$(git rev-parse --short=8 HEAD)
if [ $DRONE_BUILD_EVENT == "tag" ]; then
    build_number=$(git describe --abbrev=0 --tags)
else
    build_number=$DRONE_BUILD_NUMBER-$git_commit
fi
echo build_number
export HARBOR_BUILD_NUMBER=$build_number
upload_build=false
nightly_run=false
upload_latest_build=false
upload_bundle_success=false
latest_build_file='latest.build'
publish_npm=true

harbor_build_bundle=""
harbor_logs_bucket="harbor-ci-logs"
harbor_builds_bucket="harbor-builds"
harbor_releases_bucket="harbor-releases"
harbor_ci_pipeline_store_bucket="harbor-ci-pipeline-store/latest"
harbor_target_bucket=""
if [ $DRONE_BRANCH == "master" ]; then
  harbor_target_bucket=$harbor_builds_bucket
else
  harbor_target_bucket=$harbor_releases_bucket/$DRONE_BRANCH
fi

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
    gsutil cp $1 gs://$2/$1
    gsutil -D setacl public-read gs://$2/$1 &> /dev/null
}

function package_offline_installer {
    echo "Package Harbor offline installer."
    pybot --removekeywords TAG:secret --include Bundle tests/robot-cases/Group0-Distro-Harbor
    harbor_build_bundle=$(basename harbor-offline-installer-*.tgz)
    upload_build=true
    echo "Package name is: $harbor_build_bundle"
    du -ks $harbor_build_bundle | awk '{print $1 / 1024}' | { read x; echo $x MB; }
}

## --------------------------------------------- Run Test Case ---------------------------------------------
if [ $DRONE_REPO != "vmware/harbor" ]; then
    echo "Only run tests again Harbor Repo."
    exit 1
fi

echo "--------------------------------------------------"
echo "Running CI for $DRONE_BUILD_EVENT on $DRONE_BRANCH"
echo "--------------------------------------------------"

##
# Any merge code or tag on branch master, release-* or pks-* will trigger package offline installer.
#
# Put code here is because that it needs clean code to build installer.
##
if [ $DRONE_BRANCH == "master" ] || [ $DRONE_BRANCH == *"refs/tags"* ] || [ $DRONE_BRANCH == "release-"* ] || [ $DRONE_BRANCH == "pks-"* ]; then
    if [ $DRONE_BUILD_EVENT == "push" ] || [ $DRONE_BUILD_EVENT == "tag" ]; then
        package_offline_installer 
        upload_latest_build=true     
    fi
fi

##
# Any Event(PR or merge code) on any branch will trigger test run.
##
if (echo $buildinfo | grep -q "\[Specific CI="); then
    buildtype=$(echo $buildinfo | grep "\[Specific CI=")
    testsuite=$(echo $buildtype | awk -F"\[Specific CI=" '{sub(/\].*/,"",$2);print $2}')
    pybot -v ip:$container_ip --removekeywords TAG:secret --suite $testsuite tests/robot-cases
elif (echo $buildinfo | grep -q "\[Full CI\]"); then
    pybot -v ip:$container_ip --removekeywords TAG:secret --exclude skip tests/robot-cases
elif (echo $buildinfo | grep -q "\[Skip CI\]"); then
    echo "Skip CI."
elif (echo $buildinfo | grep -q "\[Upload Build\]"); then
    package_offline_installer
    pybot -v ip:$container_ip --removekeywords TAG:secret --include BAT tests/robot-cases/Group0-BAT
else
    # default mode is BAT.
    pybot -v ip:$container_ip --removekeywords TAG:secret --include BAT tests/robot-cases/Group0-BAT
fi

# rc is used to identify test run pass or fail.
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
#
# Build storage structure:
#
# 1(master), harbor-builds/harbor-offline-installer-*.tgz
#                         latest.build
#                         harbor-offline-installer-latest.tgz

# 2(others), harbor-releases/${branch}/harbor-offline-installer-*.tgz
#                                     latest.build
#                                     harbor-offline-installer-latest.tgz
#
if [ $upload_build == true ] && [ $rc -eq 0 ]; then
    cp $harbor_build_bundle harbor-offline-installer-latest.tgz
    uploader $harbor_build_bundle $harbor_target_bucket
    uploader harbor-offline-installer-latest.tgz $harbor_target_bucket 
    upload_bundle_success=true 
fi

## --------------------------------------------- Upload Harbor Latest Build File ----------------------------------
#
# latest.build file holds the latest offline installer url, it must be sure that the installer has been uploaded successfull.
#
if [ $upload_latest_build == true ] && [ $upload_bundle_success == true ] && [ $rc -eq 0 ]; then
    echo 'https://storage.googleapis.com/'$harbor_target_bucket/$harbor_build_bundle > $latest_build_file
    uploader $latest_build_file $harbor_target_bucket  
fi

## ------------------------------------- Build & Publish NPM Package for VIC ------------------------------------
if [ $publish_npm == true ] && [ $rc -eq 0 ] && [[ $DRONE_BUILD_EVENT == "push" || $DRONE_BUILD_EVENT == "tag" ]]; then
    echo "build & publish package harbor-ui-vic to npm repo."
    ./tools/ui_lib/build_ui_lib_4_vic.sh
fi

## ------------------------------------------------ Tear Down ---------------------------------------------------
if [ -f "$keyfile" ]; then
  rm -f $keyfile
fi

exit $rc