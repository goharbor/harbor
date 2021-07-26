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
Documentation  This resource provides any keywords related to the Harbor private registry appliance
Resource  Util.robot

*** Variables ***
${SSH_USER}  root

*** Keywords ***
Prepare Test Tools
    Wait Unitl Command Success  tar zxvf /usr/local/bin/tools.tar.gz -C /usr/local/bin/

Get And Setup Harbor CA
    [Arguments]  ${ip}  ${HARBOR_PASSWORD}  ${ca_setup_keyword}  ${ip1}==${EMPTY}
    Run Keyword If  '${ip1}' != '${EMPTY}'  Run Keywords
    ...  Get Harbor CA  ${ip1}  /drone/harbor_ca1.crt
    ...  AND  Run Keyword  ${ca_setup_keyword}  ${ip1}  ${HARBOR_PASSWORD}  /drone/harbor_ca1.crt
    Get Harbor CA  ${ip}  /drone/harbor_ca.crt
    Log To Console  ${ca_setup_keyword} ...
    Run Keyword  ${ca_setup_keyword}  ${ip}  ${HARBOR_PASSWORD}  /drone/harbor_ca.crt

Nightly Test Setup In Photon
    [Arguments]  ${ip}  ${HARBOR_PASSWORD}  ${ip1}==${EMPTY}
    Get And Setup Harbor CA  ${ip}  ${HARBOR_PASSWORD}  CA Setup In Photon  ip1=${ip1}
    Prepare Test Tools
    Log To Console  Start Docker Daemon Locally ...
    Start Docker Daemon Locally
    Log To Console  Start Containerd Daemon Locally ...
    Start Containerd Daemon Locally
    Log To Console  wget mariadb ...
    Run  wget ${prometheus_chart_file_url}
    Prepare Helm Plugin
    #Prepare docker image for push special image keyword in replication test
    Run Keyword If  '${DOCKER_USER}' != '${EMPTY}'  Docker Login  ""  ${DOCKER_USER}  ${DOCKER_PWD}

Nightly Test Setup In Ubuntu
    [Arguments]  ${ip}  ${HARBOR_PASSWORD}  ${ip1}==${EMPTY}
    Get And Setup Harbor CA  ${ip}  ${HARBOR_PASSWORD}  CA Setup In ubuntu  ip1=${ip1}
    Prepare Test Tools
    Log To Console  Start Docker Daemon Locally ...
    Run Keyword  Start Docker Daemon Locally
    Prepare Helm Plugin
    #Docker login
    Run Keyword If  '${DOCKER_USER}' != '${EMPTY}'  Docker Login  ""  ${DOCKER_USER}  ${DOCKER_PWD}

Nightly Test Setup In Ubuntu For Upgrade
    [Arguments]  ${ip}  ${HARBOR_PASSWORD}  ${ip1}==${EMPTY}
    Get And Setup Harbor CA  ${ip}  ${HARBOR_PASSWORD}  CA Setup In ubuntu  ip1=${ip1}
    Prepare Test Tools
    Log To Console  Start Docker Daemon Locally ...
    Run Keyword  Start Docker Daemon Locally
    Prepare Helm Plugin
    #For upgrade pipeline: get notary targets key from last execution.
    ${rc}  ${output}=  Run And Return Rc And Output  [ -f "/key_store/private_keys_backup.tar.gz" ] && tar -zxvf /key_store/private_keys_backup.tar.gz -C /
    #Docker login
    Run Keyword If  '${DOCKER_USER}' != '${EMPTY}'  Docker Login  ""  ${DOCKER_USER}  ${DOCKER_PWD}

CA Setup In ubuntu
    [Arguments]  ${ip}  ${HARBOR_PASSWORD}  ${cert}
    Prepare Docker Cert In Ubuntu  ${ip}  ${cert}
    #Generate Certificate Authority For Chrome  ${HARBOR_PASSWORD}

CA Setup In Photon
    [Arguments]  ${ip}  ${HARBOR_PASSWORD}  ${cert}
    Prepare Docker Cert In Photon  ${ip}  ${cert}

Collect Nightly Logs
    [Arguments]  ${ip}  ${SSH_PWD}  ${ip1}==${EMPTY}
    Run Keyword  Collect Logs  ${ip}  ${SSH_PWD}
    Run Keyword If  '${ip1}' != '${EMPTY}'  Collect Logs  ${ip1}  ${SSH_PWD}

Collect Logs
    [Arguments]  ${ip}  ${SSH_PWD}
    Open Connection    ${ip}
    Login    ${SSH_USER}    ${SSH_PWD}
    SSHLibrary.Get File  /var/log/harbor/ui.log
    SSHLibrary.Get File  /var/log/harbor/registry.log
    SSHLibrary.Get File  /var/log/harbor/proxy.log
    SSHLibrary.Get File  /var/log/harbor/adminserver.log
    SSHLibrary.Get File  /var/log/harbor/jobservice.log
    SSHLibrary.Get File  /var/log/harbor/postgresql.log
    SSHLibrary.Get File  /var/log/harbor/notary-server.log
    SSHLibrary.Get File  /var/log/harbor/notary-signer.log
    SSHLibrary.Get File  /var/log/harbor/chartmuseum.log
    SSHLibrary.Get File  /var/log/harbor/registryctl.log
    Run  rename 's/^/${ip}/' *.log
    Close All Connections
