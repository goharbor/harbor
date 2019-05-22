# Copyright Project Harbor Authors
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
    Wait Unitl Command Success  docker ${docker-params} info

Pull image
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}
    Log To Console  \nRunning docker pull ${image}...
    Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    ${output}=  Wait Unitl Command Success  docker pull ${ip}/${project}/${image}
    Should Contain  ${output}  Digest:
    Should Contain  ${output}  Status:
    Should Not Contain  ${output}  No such image:

Push image
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}
    Log To Console  \nRunning docker push ${image}...
    Wait Unitl Command Success  docker pull ${image}
    Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    Wait Unitl Command Success  docker tag ${image} ${ip}/${project}/${image}
    Wait Unitl Command Success  docker push ${ip}/${project}/${image}
    Wait Unitl Command Success  docker logout ${ip}

Push Image With Tag
#tag1 is tag of image on docker hub,default latest,use a version existing if you do not want to use latest    
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}  ${tag}  ${tag1}=latest
    Log To Console  \nRunning docker push ${image}...
    Wait Unitl Command Success  docker pull ${image}:${tag1}
    Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    Wait Unitl Command Success  docker tag ${image}:${tag1} ${ip}/${project}/${image}:${tag}
    Wait Unitl Command Success  docker push ${ip}/${project}/${image}:${tag}
    Wait Unitl Command Success  docker logout ${ip}

Cannot Pull image
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}
    Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    Wait Unitl Command Success  docker pull ${ip}/${project}/${image}  positive=${false}

Cannot Pull Unsigned Image
    [Arguments]  ${ip}  ${user}  ${pass}  ${proj}  ${imagewithtag}  
    Wait Unitl Command Success  docker login -u ${user} -p ${pass} ${ip}
    ${output}=  Wait Unitl Command Success  docker pull ${ip}/${proj}/${imagewithtag}  positive=${false}
    Should Contain  ${output}  The image is not signed in Notary

Cannot Push image
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}
    Log To Console  \nRunning docker push ${image}...
    Wait Unitl Command Success  docker pull ${image}
    Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    Wait Unitl Command Success  docker tag ${image} ${ip}/${project}/${image}
    Wait Unitl Command Success  docker push ${ip}/${project}/${image}  positive=${false}
    Wait Unitl Command Success  docker logout ${ip}

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
    Wait Unitl Command Success  wget ${vch-ip}:${port}

Get Container IP
    [Arguments]  ${docker-params}  ${id}  ${network}=default  ${dockercmd}=docker
    ${ip}=  Wait Unitl Command Success  ${dockercmd} ${docker-params} network inspect ${network} | jq '.[0].Containers."${id}".IPv4Address' | cut -d \\" -f 2 | cut -d \\/ -f 1
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
    Wait Unitl Command Success  mkdir -p /etc/docker/certs.d/${ip}
    Wait Unitl Command Success  cp harbor_ca.crt /etc/docker/certs.d/${ip}

Kill Local Docker Daemon
    [Arguments]  ${handle}  ${dockerd-pid}
    Terminate Process  ${handle}
    Process Should Be Stopped  ${handle}
    Wait Unitl Command Success  kill -9 ${dockerd-pid}

Docker Login Fail
    [Arguments]  ${ip}  ${user}  ${pwd}
    Log To Console  \nRunning docker login ${ip} ...
    ${output}=  Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}  positive=${false}
    Should Contain  ${output}  unauthorized: authentication required
    Should Not Contain  ${output}  500 Internal Server Error

Docker Login
    [Arguments]  ${server}  ${username}  ${password}
    Wait Unitl Command Success  docker login -u ${username} -p ${password} ${server}

Docker Pull
    [Arguments]  ${image}
    Wait Unitl Command Success  docker pull ${image}

Docker Tag
    [Arguments]  ${src_image}   ${dst_image}
    Wait Unitl Command Success  docker tag ${src_image} ${dst_image}

Docker Push
    [Arguments]  ${image}
    Wait Unitl Command Success  docker push ${image}