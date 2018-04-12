# Copyright 2016-2017 VMware, Inc. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License

*** Settings ***
Documentation  This resource provides helper functions for docker operations
Library  OperatingSystem
Library  Process

*** Keywords ***
Run Docker Info
    [Arguments]  ${docker-params}
    ${rc}=  Run And Return Rc  docker ${docker-params} info
    Should Be Equal As Integers  ${rc}  0

Pull image
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}
    Log To Console  \nRunning docker pull ${image}...
    ${rc}  ${output}=  Run And Return Rc And Output  docker login -u ${user} -p ${pwd} ${ip}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker pull ${ip}/${project}/${image}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  Digest:
    Should Contain  ${output}  Status:
    Should Not Contain  ${output}  No such image:

Push image
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}
    Log To Console  \nRunning docker push ${image}...
    ${rc}  ${output}=  Run And Return Rc And Output  docker pull ${image}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker login -u ${user} -p ${pwd} ${ip}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  docker tag ${image} ${ip}/${project}/${image}
    ${rc}  ${output}=  Run And Return Rc And Output  docker push ${ip}/${project}/${image}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  docker logout ${ip}

Push Image With Tag
#tag1 is tag of image on docker hub,default latest,use a version existing if you do not want to use latest    
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}  ${tag}  ${tag1}=latest
    Log To Console  \nRunning docker push ${image}...
    ${rc}  ${output}=  Run And Return Rc And Output  docker pull ${image}:${tag1}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker login -u ${user} -p ${pwd} ${ip}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  docker tag ${image}:${tag1} ${ip}/${project}/${image}:${tag}
    ${rc}  ${output}=  Run And Return Rc And Output  docker push ${ip}/${project}/${image}:${tag}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  docker logout ${ip}

Cannot Pull image
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}
    ${rc}  ${output}=  Run And Return Rc And Output  docker login -u ${user} -p ${pwd} ${ip}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker pull ${ip}/${project}/${image}
    Log  ${output}
    Should Not Be Equal As Integers  ${rc}  0

Cannot Pull Unsigned Image
    [Arguments]  ${ip}  ${user}  ${pass}  ${proj}  ${imagewithtag}  
    ${rc}  ${output}=  Run And Return Rc And Output  docker login -u ${user} -p ${pass} ${ip}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker pull ${ip}/${proj}/${imagewithtag}
    Should Contain  ${output}  The image is not signed in Notary
    Should Not Be Equal As Integers  ${rc}  0

Cannot Push image
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}
    Log To Console  \nRunning docker push ${image}...
    ${rc}=  Run And Return Rc  docker pull ${image}
    ${rc}  ${output}=  Run And Return Rc And Output  docker login -u ${user} -p ${pwd} ${ip}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  docker tag ${image} ${ip}/${project}/${image}
    ${rc}  ${output}=  Run And Return Rc And Output  docker push ${ip}/${project}/${image}
    Log  ${output}
    Should Not Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  docker logout ${ip}

Wait Until Container Stops
    [Arguments]  ${container}
    :FOR  ${idx}  IN RANGE  0  60
    \   ${out}=  Run  docker %{VCH-PARAMS} inspect ${container} | grep Status
    \   ${status}=  Run Keyword And Return Status  Should Contain  ${out}  exited
    \   Return From Keyword If  ${status}
    \   Sleep  1
    Fail  Container did not stop within 60 seconds

Hit Nginx Endpoint
    [Arguments]  ${vch-ip}  ${port}
    ${rc}  ${output}=  Run And Return Rc And Output  wget ${vch-ip}:${port}
    Should Be Equal As Integers  ${rc}  0

Get Container IP
    [Arguments]  ${docker-params}  ${id}  ${network}=default  ${dockercmd}=docker
    ${rc}  ${ip}=  Run And Return Rc And Output  ${dockercmd} ${docker-params} network inspect ${network} | jq '.[0].Containers."${id}".IPv4Address' | cut -d \\" -f 2 | cut -d \\/ -f 1
    Should Be Equal As Integers  ${rc}  0
    [Return]  ${ip}

# The local dind version is embedded in Dockerfile
# docker:1.13-dind
# If you are running this keyword in a container, make sure it is run with --privileged turned on
Start Docker Daemon Locally
    ${pid}=  Run  pidof dockerd
    Return From Keyword If  '${pid}' != '${EMPTY}'
    OperatingSystem.File Should Exist  /usr/local/bin/dockerd-entrypoint.sh
    ${handle}=  Start Process  /usr/local/bin/dockerd-entrypoint.sh dockerd>./daemon-local.log 2>&1  shell=True
    Process Should Be Running  ${handle}
    :FOR  ${IDX}  IN RANGE  5
    \   ${pid}=  Run  pidof dockerd
    \   Exit For Loop If  '${pid}' != '${EMPTY}'
    \   Sleep  2s
    Sleep  2s
    [Return]  ${handle}

Prepare Docker Cert
    [Arguments]  ${ip}
    ${rc}  ${out}=  Run And Return Rc And Output  mkdir -p /etc/docker/certs.d/${ip}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${out}=  Run And Return Rc And Output  cp harbor_ca.crt /etc/docker/certs.d/${ip}
    Should Be Equal As Integers  ${rc}  0   
    
Kill Local Docker Daemon
    [Arguments]  ${handle}  ${dockerd-pid}
    Terminate Process  ${handle}
    Process Should Be Stopped  ${handle}
    ${rc}=  Run And Return Rc  kill -9 ${dockerd-pid}
    Should Be Equal As Integers  ${rc}  0

Docker Login Fail
    [Arguments]  ${ip}  ${user}  ${pwd}
    Log To Console  \nRunning docker login ${ip} ...
    ${rc}  ${output}=  Run And Return Rc And Output  docker login -u ${user} -p ${pwd} ${ip}
    Should Not Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  unauthorized: authentication required
    Should Not Contain  ${output}  500 Internal Server Error
