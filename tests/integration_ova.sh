#!/bin/bash
# Copyright 2017 VMware, Inc. All Rights Reserved.
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
container_ip=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
echo $container_ip

ova_url="$(python /auto-ova/ova.py)"
echo $ova_url

## --------------------------------------------- Init Env -------------------------------------------------
# Start Xvfb for Chrome headlesss
Xvfb -ac :99 -screen 0 1280x1024x16 & export DISPLAY=:99

## --------------------------------------------- Run -------------------------------------------------
pybot -v ip:$container_ip -v ova_url:$ova_url --include OVA tests/robot-cases/Group5-OVA-install-config/5-00-OVA-BAT.robot

## --------------------------------------------- Tear Down -------------------------------------------------
rc="$?"
echo $rc
exit $rc