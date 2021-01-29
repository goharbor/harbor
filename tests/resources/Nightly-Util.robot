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
Nightly Test Setup
    [Arguments]  ${ip}  ${HARBOR_PASSWORD}  ${ip1}==${EMPTY}
    Run Keyword If  '${ip1}' != '${EMPTY}'  CA setup  ${ip1}  ${HARBOR_PASSWORD}  /ca/ca1.crt
    Run Keyword If  '${ip1}' != '${EMPTY}'  Run  rm -rf ./harbor_ca.crt
    Log To Console  CA setup ...
    Run Keyword  CA setup  ${ip}  ${HARBOR_PASSWORD}
    Log To Console  Start Docker Daemon Locally ...
    Run Keyword  Start Docker Daemon Locally
    Log To Console  Start Containerd Daemon Locally ...
    Run Keyword  Start Containerd Daemon Locally
    Log To Console  wget mariadb ...
    Run  wget ${prometheus_chart_file_url}
    #Prepare docker image for push special image keyword in replication test
    Run Keyword If  '${DOCKER_USER}' != '${EMPTY}'  Docker Login  ""  ${DOCKER_USER}  ${DOCKER_PWD}

CA Setup
    [Arguments]  ${ip}  ${HARBOR_PASSWORD}  ${cert}=/ca/ca.crt
    Log To Console  cp /ca/harbor_ca.crt harbor_ca.crt ...
    Run  cp /ca/harbor_ca.crt harbor_ca.crt
    Log To Console  Generate Certificate Authority For Chrome ...
    Generate Certificate Authority For Chrome  ${HARBOR_PASSWORD}
    Log To Console  Prepare Docker Cert ...
    Prepare Docker Cert  ${ip}

Nightly Test Setup For Nightly
    [Arguments]  ${ip}  ${HARBOR_PASSWORD}  ${ip1}==${EMPTY}
    Run Keyword If  '${ip1}' != '${EMPTY}'  CA setup For Nightly  ${ip1}  ${HARBOR_PASSWORD}  /ca/ca1.crt
    Run Keyword If  '${ip1}' != '${EMPTY}'  Run  rm -rf ./harbor_ca.crt
    Run Keyword  CA setup For Nightly  ${ip}  ${HARBOR_PASSWORD}
    Log To Console  Start Docker Daemon Locally ...
    Run Keyword  Start Docker Daemon Locally
    Log To Console  Start Containerd Daemon Locally ...
    Run Keyword  Start Containerd Daemon Locally
    #Prepare docker image for push special image keyword in replication test
    Docker Pull  ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/busybox:latest
    Docker Tag  ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/busybox:latest  busybox:latest

CA Setup For Nightly
    [Arguments]  ${ip}  ${HARBOR_PASSWORD}  ${cert}=/ca/ca.crt
    Run  cp ${cert} harbor_ca.crt
    Generate Certificate Authority For Chrome  ${HARBOR_PASSWORD}
    Prepare Docker Cert For Nightly  ${ip}
    Prepare Helm Cert

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