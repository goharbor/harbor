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
Documentation  This resource provides any keywords related to the Harbor private registry appliance
Resource  Util.robot

*** Variables ***
${SSH_USER}  root

*** Keywords ***
Nightly Test Setup
    [Arguments]  ${ip}  ${SSH_PWD}  ${HARBOR_PASSWORD}  ${ip1}==${EMPTY}
    Run Keyword  CA setup  ${ip}  ${SSH_PWD}  ${HARBOR_PASSWORD}
    Run Keyword  Prepare Docker Cert  ${ip}    
    Run Keyword If  '${ip1}' != '${EMPTY}'  Run  rm harbor_ca.crt 
    Run Keyword If  '${ip1}' != '${EMPTY}'  CA setup  ${ip1}  ${SSH_PWD}  ${HARBOR_PASSWORD}
    Run Keyword If  '${ip1}' != '${EMPTY}'  Prepare Docker Cert  ${ip1}
    Run Keyword  Start Docker Daemon Locally

CA Setup
    [Arguments]  ${ip}  ${SSH_PWD}  ${HARBOR_PASSWORD}
    Open Connection    ${ip}
    Login    ${SSH_USER}    ${SSH_PWD}
    SSHLibrary.Get File  /data/ca_download/ca.crt
    Close All Connections
    Run  mv ca.crt harbor_ca.crt
    Generate Certificate Authority For Chrome  ${HARBOR_PASSWORD}	

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
    SSHLibrary.Get File  /var/log/harbor/clair.log  
    SSHLibrary.Get File  /var/log/harbor/jobservice.log  
    SSHLibrary.Get File  /var/log/harbor/postgresql.log
    SSHLibrary.Get File  /var/log/harbor/notary-server.log
    SSHLibrary.Get File  /var/log/harbor/notary-signer.log
    Run  rename 's/^/${ip}/' *.log
    Close All Connections