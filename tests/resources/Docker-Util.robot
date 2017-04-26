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
    [Arguments]  ${image}
    Log To Console  \nRunning docker pull ${image}...
    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} pull ${image}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  Digest:
    Should Contain  ${output}  Status:
    Should Not Contain  ${output}  No such image:

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
    [Arguments]  ${dockerd-params}  ${dockerd-path}=/usr/local/bin/dockerd-entrypoint.sh  ${log}=./daemon-local.log
    OperatingSystem.File Should Exist  ${dockerd-path}
    ${handle}=  Start Process  ${dockerd-path} ${dockerd-params} >${log} 2>&1  shell=True
    Process Should Be Running  ${handle}
    :FOR  ${IDX}  IN RANGE  5
    \   ${pid}=  Run  pidof dockerd
    \   Run Keyword If  '${pid}' != '${EMPTY}'  Set Test Variable  ${dockerd-pid}  ${pid}
    \   Exit For Loop If  '${pid}' != '${EMPTY}'
    \   Sleep  1s
    Should Not Be Equal  '${dockerd-pid}'  '${EMPTY}'
    [Return]  ${handle}  ${dockerd-pid}

Kill Local Docker Daemon
    [Arguments]  ${handle}  ${dockerd-pid}
    Terminate Process  ${handle}
    Process Should Be Stopped  ${handle}
    ${rc}=  Run And Return Rc  kill -9 ${dockerd-pid}
    Should Be Equal As Integers  ${rc}  0

Get container shortID
    [Arguments]  ${id}
    ${shortID}=  Get Substring  ${id}  0  12
    [Return]  ${shortID}

Get VM display name
    [Arguments]  ${id}
    ${rc}  ${name}=  Run And Return Rc And Output  docker %{VCH-PARAMS} inspect --format='{{.Name}}' ${id}
    Should Be Equal As Integers  ${rc}  0
    ${name}=  Get Substring  ${name}  1
    ${shortID}=  Get container shortID  ${id}
    [Return]  ${name}-${shortID}

Verify Container Rename
    [Arguments]  ${oldname}  ${newname}  ${contID}
    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} ps -a
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  ${newname}
    Should Not Contain  ${output}  ${oldname}
    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} inspect -f '{{.Name}}' ${newname}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  ${newname}
    ${vmName}=  Get VM display name  ${contID}
    ${rc}  ${output}=  Run And Return Rc And Output  govc vm.info ${vmname}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  ${vmName}

Run Regression Tests
    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} pull busybox
    Should Be Equal As Integers  ${rc}  0
    # Pull an image that has been pulled already
    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} pull busybox
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} images
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  busybox
    ${rc}  ${container}=  Run And Return Rc And Output  docker %{VCH-PARAMS} create busybox /bin/top
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} start ${container}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} ps
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  /bin/top
    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} stop ${container}
    Should Be Equal As Integers  ${rc}  0
    Wait Until Container Stops  ${container}
    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} ps -a
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  Exited

    ${vmName}=  Get VM Display Name  ${container}
    Wait Until Keyword Succeeds  5x  10s  Check For The Proper Log Files  ${vmName}

    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} rm ${container}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} ps -a
    Should Be Equal As Integers  ${rc}  0
    Should Not Contain  ${output}  /bin/top

    # Check for regression for #1265
    ${rc}  ${container1}=  Run And Return Rc And Output  docker %{VCH-PARAMS} create -it busybox /bin/top
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${container2}=  Run And Return Rc And Output  docker %{VCH-PARAMS} create -it busybox
    Should Be Equal As Integers  ${rc}  0
    ${shortname}=  Get Substring  ${container2}  1  12
    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} ps -a
    ${lines}=  Get Lines Containing String  ${output}  ${shortname}
    Should Not Contain  ${lines}  /bin/top
    ${rc}=  Run And Return Rc  docker %{VCH-PARAMS} rm ${container1}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  docker %{VCH-PARAMS} rm ${container2}
    Should Be Equal As Integers  ${rc}  0

    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} rmi busybox
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker %{VCH-PARAMS} images
    Should Be Equal As Integers  ${rc}  0
    Should Not Contain  ${output}  busybox

    Scrape Logs For The Password

Launch Container
    [Arguments]  ${name}  ${network}=default  ${dockercmd}=docker
    ${rc}  ${output}=  Run And Return Rc And Output  ${dockercmd} %{VCH-PARAMS} run --name ${name} --net ${network} -itd busybox
    Should Be Equal As Integers  ${rc}  0
    ${id}=  Get Line  ${output}  -1
    ${ip}=  Get Container IP  %{VCH-PARAMS}  ${id}  ${network}  ${dockercmd}
    [Return]  ${id}  ${ip}
