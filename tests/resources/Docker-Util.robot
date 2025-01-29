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
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}  ${tag}=${null}  ${is_robot}=${false}
    Log To Console  \nRunning docker pull ${image}...
    ${image_with_tag}=  Set Variable If  '${tag}'=='${null}'  ${image}  ${image}:${tag}
    Run Keyword If  ${is_robot}==${false}  Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    ...  ELSE  Wait Unitl Command Success  docker login -u robot\\\$${project}+${user} -p ${pwd} ${ip}
    ${output}=  Docker Pull  ${ip}/${project}/${image_with_tag}
    Log  ${output}
    Log To Console  ${output}
    Should Contain  ${output}  Digest:
    Should Contain  ${output}  Status:
    Should Not Contain  ${output}  No such image:

Push image
    # If no tag provided in $(image_with_or_without_tag}, latest will be the tag pulled from docker-hub or read from local
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image_with_or_without_tag}  ${need_pull_first}=${true}  ${sha256}=${null}  ${is_robot}=${false}
    ${d}=    Get Current Date    result_format=%m%s
    ${image_in_use}=  Set Variable If  '${sha256}'=='${null}'  ${image_with_or_without_tag}  ${image_with_or_without_tag}@sha256:${sha256}
    ${image_in_use_with_tag}=  Set Variable If  '${sha256}'=='${null}'  ${image_with_or_without_tag}  ${image_with_or_without_tag}:${sha256}
    Sleep  3
    Log To Console  \nRunning docker push ${image_with_or_without_tag}...
    ${image_in_use}=   Set Variable If  ${need_pull_first}==${true}  ${image_in_use}  ${image_with_or_without_tag}
    Run Keyword If  ${need_pull_first}==${true}   Docker Pull  ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image_in_use}
    Run Keyword If  ${is_robot}==${false}  Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    ...  ELSE  Wait Unitl Command Success  docker login -u robot\\\$${project}+${user} -p ${pwd} ${ip}
    Run Keyword If  ${need_pull_first}==${true}  Wait Unitl Command Success  docker tag ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image_in_use} ${ip}/${project}/${image_in_use_with_tag}
    ...  ELSE  Wait Unitl Command Success  docker tag ${image_in_use} ${ip}/${project}/${image_in_use_with_tag}
    Wait Unitl Command Success  docker push ${ip}/${project}/${image_in_use_with_tag}
    Wait Unitl Command Success  docker logout ${ip}
    Sleep  1

Push Image With Tag
#tag1 is tag of image on docker hub,default latest,use a existed version if you do not want to use latest
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}  ${tag}  ${tag1}=latest
    Log To Console  \nRunning docker push ${image}...
    Docker Pull  ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}:${tag1}
    Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    Wait Unitl Command Success  docker tag ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}:${tag1} ${ip}/${project}/${image}:${tag}
    Wait Unitl Command Success  docker push ${ip}/${project}/${image}:${tag}
    Wait Unitl Command Success  docker logout ${ip}

Cannot Docker Login Harbor
    [Arguments]  ${ip}  ${user}  ${pwd}
    Command Should be Failed  docker login -u ${user} -p ${pwd} ${ip}

Cannot Pull Image
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}  ${tag}=${null}  ${err_msg}=${null}
    ${image_with_tag}=  Set Variable If  '${tag}'=='${null}'  ${image}  ${image}:${tag}
    Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    FOR  ${idx}  IN RANGE  0  30
        ${out}  Run Keyword And Ignore Error  Command Should be Failed  docker pull ${ip}/${project}/${image_with_tag}
        Exit For Loop If  '${out[0]}'=='PASS'
        Sleep  3
    END
    Log To Console  Cannot Pull Image - Pull Log: ${out[1]}
    Should Be Equal As Strings  '${out[0]}'  'PASS'
    Run Keyword If  '${err_msg}' != '${null}'  Should Contain  ${out[1]}  ${err_msg}

Cannot Push image
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${image}  ${err_msg}=${null}  ${err_msg_2}=${null}
    Log To Console  \nRunning docker push ${image}...
    Docker Pull  ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}
    Wait Unitl Command Success  docker login -u ${user} -p ${pwd} ${ip}
    Wait Unitl Command Success  docker tag ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image} ${ip}/${project}/${image}
    ${output}=  Command Should be Failed  docker push ${ip}/${project}/${image}
    Log To Console  ${output}
    Run Keyword If  '${err_msg}' != '${null}'  Should Contain  ${output}  ${err_msg}
    Run Keyword If  '${err_msg_2}' != '${null}'  Should Contain  ${output}  ${err_msg_2}
    Wait Unitl Command Success  docker logout ${ip}

Wait Until Container Stops
    [Arguments]  ${container}
    FOR  ${idx}  IN RANGE  0  60
        ${out}=  Run  docker %{VCH-PARAMS} inspect ${container} | grep Status
        ${status}=  Run Keyword And Return Status  Should Contain  ${out}  exited
        Return From Keyword If  ${status}
        Sleep  1
    END
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
    #${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group0-Util/docker_config.sh
    #Log  ${output}
    #Should Be Equal As Integers  ${rc}  0
    Return From Keyword If  '${pid}' != '${EMPTY}'
    OperatingSystem.File Should Exist  /usr/local/bin/dockerd-entrypoint.sh
    ${handle}=  Start Process  /usr/local/bin/dockerd-entrypoint.sh dockerd>./daemon-local.log 2>&1  shell=True
    Process Should Be Running  ${handle}
    FOR  ${IDX}  IN RANGE  5
        ${pid}=  Run  pidof dockerd
        Exit For Loop If  '${pid}' != '${EMPTY}'
        Sleep  2s
    END
    Sleep  2s
    [Return]  ${handle}

Start Containerd Daemon Locally
    ${handle}=  Start Process  /usr/local/bin/containerd > ./daemon-local.log 2>&1 &  shell=True
    FOR  ${IDX}  IN RANGE  5
        ${pid}=  Run  pidof /usr/local/bin/containerd
        Log To Console  pid: ${pid}
        Exit For Loop If  '${pid}' != '${EMPTY}'
        Sleep  2s
    END
    Sleep  2s
    [Return]  ${handle}

Restart Process Locally
    [Arguments]  ${process}
    ${full_process}=  Set Variable If
    ...  '${process}'=='containerd'  /usr/local/bin/containerd  dockerd
    ${start_process_cmd}=  Set Variable If
    ...  '${process}'=='dockerd'  /usr/local/bin/dockerd-entrypoint.sh dockerd>./daemon-local.log 2>&1
    ...  '${process}'=='containerd'  ${full_process} > ./daemon-local.log 2>&1 &
    Should Be True  '${start_process_cmd}' != '${EMPTY}'
    Run Keyword If  '${process}'=='dockerd'  OperatingSystem.File Should Exist  /usr/local/bin/dockerd-entrypoint.sh

    FOR  ${IDX}  IN RANGE  5
        ${pid}=  Run  pidof ${full_process}
        Exit For Loop If  '${pid}' == '${EMPTY}'
        ${result}=  Run  kill ${pid}
        Log To Console  Kill docker process: ${result}
        Sleep  2s
    END
    ${pid}=  Run  pidof ${full_process}
    Should Be Equal As Strings  '${pid}'  '${EMPTY}'

    ${result}=  Run  rm -rf /var/lib/${process}/*
    Log All  Clear /var/lib/${process}: ${result}
    ${handle}=  Start Process  ${start_process_cmd}  shell=True
    Log All  handle : ${handle}
    FOR  ${IDX}  IN RANGE  5
        ${pid}=  Run  pidof ${full_process}
        Log All  pid : ${pid}
        Exit For Loop If  '${pid}' != '${EMPTY}'
        Sleep  2s
    END
    Sleep  2s
    #Process Should Be Running  ${handle}
    ${result}=  Run  ps aux |grep ${full_process}
    Log All  result : ${result}
    [Return]  ${handle}

Prepare Docker Cert In Ubuntu
    [Arguments]  ${ip}  ${cert}
    Wait Unitl Command Success  rm -rf ~/.docker/
    Wait Unitl Command Success  mkdir -p /etc/docker/certs.d/${ip}
    Wait Unitl Command Success  cp ${cert} /etc/docker/certs.d/${ip}
    Wait Unitl Command Success  cp ${cert} /usr/local/share/ca-certificates/
    #Add pivotal ecs cert for docker manifest push test.
    Wait Unitl Command Success  cp /ecs_ca/vmwarecert.crt /usr/local/share/ca-certificates/
    Wait Unitl Command Success  update-ca-certificates

Prepare Docker Cert In Photon
    [Arguments]  ${ip}  ${cert}
    Log All  Prepare Docker Cert In Photon ${cert}
    ${rc}  ${output}=  Run And Return Rc and Output  cat ${cert}
    Log All  CA output: ${output}
    Wait Unitl Command Success  cat ${cert} >> /etc/pki/tls/certs/ca-bundle.crt
    Wait Unitl Command Success  mkdir -p /etc/docker/certs.d/${ip}
    Wait Unitl Command Success  cp ${cert} /etc/docker/certs.d/${ip}

Kill Local Docker Daemon
    [Arguments]  ${handle}  ${dockerd-pid}
    Terminate Process  ${handle}
    Process Should Be Stopped  ${handle}
    Wait Unitl Command Success  kill -9 ${dockerd-pid}

Clean All Local Images
    ${rc}  ${out}=  Run Keyword And Ignore Error  Run  docker rmi -f $(docker images -a -q)
    Log All  ${out}
    ${rc}  ${out}=  Run Keyword And Ignore Error  Run  docker system prune -a -f
    Log All  ${out}

Docker Login Fail
    [Arguments]  ${ip}  ${user}  ${pwd}
    Log To Console  \nRunning docker login ${ip} ...
    ${prev_lvl}  Set Log Level  NONE
    ${output}=  Command Should be Failed  docker login -u ${user} -p ${pwd} ${ip}
    ${prev_lvl}  Set Log Level  ${prev_lvl}
    Should Contain  ${output}  unauthorized
    Should Not Contain  ${output}  500 Internal Server Error

Docker Login
    [Arguments]  ${server}  ${username}  ${password}
    ${prev_lvl}  Set Log Level  NONE
    Wait Unitl Command Success  docker login -u ${username} -p ${password} ${server}
    ${prev_lvl}  Set Log Level  ${prev_lvl}

Docker Logout
    [Arguments]  ${server}
    Wait Unitl Command Success  docker logout ${server}

Docker Pull
    [Arguments]  ${image}
    ${output}=  Retry Keyword N Times When Error  6  Wait Unitl Command Success  docker pull ${image}
    Log All  Docker Pull: ${output}
    [Return]  ${output}

Docker Tag
    [Arguments]  ${src_image}   ${dst_image}
    Wait Unitl Command Success  docker tag ${src_image} ${dst_image}

Docker Push
    [Arguments]  ${image}
    Wait Unitl Command Success  docker push ${image}

Docker Push Index
    [Arguments]  ${ip}  ${user}  ${pwd}  ${index}  ${image1}  ${image2}
    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group0-Util/docker_push_manifest_list.sh ${ip} ${user} ${pwd} ${index} ${image1} ${image2}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0

Docker Image Can Not Be Pulled
    [Arguments]  ${image}
    FOR  ${idx}  IN RANGE  0  30
        ${out}=  Run Keyword And Ignore Error  Docker Login  ""  ${DOCKER_USER}  ${DOCKER_PWD}
        Log To Console  Return value is ${out}
        ${out}=  Run Keyword And Ignore Error  Command Should be Failed  docker pull ${image}
        Exit For Loop If  '${out[0]}'=='PASS'
        Log To Console  Docker pull return value is ${out}
        Sleep  3
    END
    Log To Console  Cannot Pull Image From Docker - Pull Log: ${out[1]}
    Should Be Equal As Strings  '${out[0]}'  'PASS'

Docker Image Can Not Be Pulled With Credential
    [Arguments]  ${server}  ${username}  ${password}  ${image}
    FOR  ${idx}  IN RANGE  0  30
        ${out}=  Run Keyword And Ignore Error  Docker Login  ${server}  ${username}  ${password}
        Log To Console  Return value is ${out}
        ${out}=  Run Keyword And Ignore Error  Command Should be Failed  docker pull ${image}
        Exit For Loop If  '${out[0]}'=='PASS'
        Log To Console  Docker pull return value is ${out}
        Sleep  3
    END
    Log To Console  Cannot Pull Image From Docker - Pull Log: ${out[1]}
    Should Be Equal As Strings  '${out[0]}'  'PASS'

Docker Image Can Be Pulled
    [Arguments]  ${image}  ${period}=60  ${times}=2
    FOR  ${n}  IN RANGE  1  ${times}
        Sleep  ${period}
        ${out}=  Run Keyword And Ignore Error  Docker Login  ""  ${DOCKER_USER}  ${DOCKER_PWD}
        Log To Console  Return value is ${out}
        ${out}=  Run Keyword And Ignore Error  Docker Pull  ${image}
        Log To Console  Return value is ${out[0]}
        Exit For Loop If  '${out[0]}'=='PASS'
        Sleep  5
    END
    Run Keyword If  '${out[0]}'=='FAIL'  Capture Page Screenshot
    Should Be Equal As Strings  '${out[0]}'  'PASS'

Docker Image Can Be Pulled With Credential
    [Arguments]  ${server}  ${username}  ${password}  ${image}  ${period}=60  ${times}=2
    FOR  ${n}  IN RANGE  1  ${times}
        Sleep  ${period}
        ${out}=  Run Keyword And Ignore Error  Docker Login  ${server}  ${username}  ${password}
        Log To Console  Return value is ${out}
        ${out}=  Run Keyword And Ignore Error  Docker Pull  ${image}
        Log To Console  Return value is ${out[0]}
        Exit For Loop If  '${out[0]}'=='PASS'
        Sleep  5
    END
    Run Keyword If  '${out[0]}'=='FAIL'  Capture Page Screenshot
    Should Be Equal As Strings  '${out[0]}'  'PASS'