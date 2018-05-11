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
Library  Selenium2Library
Library  OperatingSystem

*** Variables ***
${HARBOR_VERSION}  v1.1.1
${CLAIR_BUILDER}  1.4.1
${GOLANG_VERSION}  1.9.2

*** Keywords ***
Install Harbor to Test Server
    Log To Console  \nStart Docker Daemon
    Start Docker Daemon Locally
    Sleep  5s
    ${rc}  ${output}=  Run And Return Rc And Output  docker ps
    Should Be Equal As Integers  ${rc}  0
    Log To Console  \n${output}
    Log To Console  \nconfig harbor cfg
    Config Harbor cfg  http_proxy=https
    Prepare Cert
    Log To Console  \ncomplile and up harbor now
    Compile and Up Harbor With Source Code
    ${rc}  ${output}=  Run And Return Rc And Output  docker ps
    Should Be Equal As Integers  ${rc}  0
    Log To Console  \n${output}
    Generate Certificate Authority For Chrome

Up Harbor
    [Arguments]  ${with_notary}=true  ${with_clair}=true
    ${rc}  ${output}=  Run And Return Rc And Output  make start -e NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair}
    Log  ${rc}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0

Down Harbor
    [Arguments]  ${with_notary}=true  ${with_clair}=true
    ${rc}  ${output}=  Run And Return Rc And Output  echo "Y" | make down -e NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair}
    Log  ${rc}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0

Package Harbor Offline
    [Arguments]  ${golang_image}=golang:${GOLANG_VERSION}  ${clarity_image}=vmware/harbor-clarity-ui-builder:${CLAIR_BUILDER}  ${with_notary}=true  ${with_clair}=true  ${with_migrator}=true
    Log To Console  \nStart Docker Daemon
    Start Docker Daemon Locally
    ${rc}  ${output}=  Run And Return Rc And Output  make package_offline VERSIONTAG=%{Harbor_Assets_Version} PKGVERSIONTAG=%{Harbor_Package_Version} UIVERSIONTAG=%{Harbor_UI_Version} GOBUILDIMAGE=${golang_image} COMPILETAG=compile_golangimage CLARITYIMAGE=${clarity_image} NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair} MIGRATORFLAG=${with_migrator} HTTPPROXY=
    Log  ${rc}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0

Package Harbor Online
    [Arguments]  ${golang_image}=golang:${GOLANG_VERSION}  ${clarity_image}=vmware/harbor-clarity-ui-builder:${CLAIR_BUILDER}  ${with_notary}=true  ${with_clair}=true  ${with_migrator}=true
    Log To Console  \nStart Docker Daemon
    Start Docker Daemon Locally
    ${rc}  ${output}=  Run And Return Rc And Output  make package_online VERSIONTAG=%{Harbor_Assets_Version} PKGVERSIONTAG=%{Harbor_Package_Version} UIVERSIONTAG=%{Harbor_UI_Version} GOBUILDIMAGE=${golang_image} COMPILETAG=compile_golangimage CLARITYIMAGE=${clarity_image} NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair} MIGRATORFLAG=${with_migrator} HTTPPROXY=
    Log  ${rc}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    
Switch To LDAP
    Down Harbor
    ${rc}  ${output}=  Run And Return Rc And Output  rm -rf /data
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0
    Prepare Cert
    Config Harbor cfg  auth=ldap_auth  http_proxy=https
    Prepare
    Up Harbor
    ${rc}=  Run And Return Rc  docker pull osixia/openldap:1.1.7
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  cd tests && ./ldapprepare.sh
    Log  ${rc}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker ps
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0	
    Generate Certificate Authority For Chrome
	
Enable Notary Client
    ${rc}  ${output}=  Run And Return Rc And Output  rm -rf ~/.docker/
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0
    Log  ${ip}
    ${rc}=  Run And Return Rc  mkdir -p /etc/docker/certs.d/${ip}/
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  mkdir -p ~/.docker/tls/${ip}:4443/
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  cp ./harbor_ca.crt /etc/docker/certs.d/${ip}/
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  cp ./harbor_ca.crt ~/.docker/tls/${ip}:4443/
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  ls -la /etc/docker/certs.d/${ip}/
    Log  ${output}
    ${rc}  ${output}=  Run And Return Rc And Output  ls -la ~/.docker/tls/${ip}:4443/
    Log  ${output}

Prepare
    [Arguments]  ${with_notary}=true  ${with_clair}=true
    ${rc}  ${output}=  Run And Return Rc And Output  make prepare -e NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair}
    Log  ${rc}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0

Config Harbor cfg
    # Will change the IP and Protocol in the harbor.cfg
    [Arguments]  ${http_proxy}=http  ${auth}=db_auth
    ${rc}  ${output}=  Run And Return Rc And Output  ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'
    ${rc}=  Run And Return Rc  sed "s/^hostname = .*/hostname = ${output}/g" -i ./make/harbor.cfg
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  sed "s/^ui_url_protocol = .*/ui_url_protocol = ${http_proxy}/g" -i ./make/harbor.cfg
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  sed "s/^auth_mode = .*/auth_mode = ${auth}/g" -i ./make/harbor.cfg
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0
    ${out}=  Run  cat ./make/harbor.cfg
    Log  ${out}		

Prepare Cert
    # Will change the IP and Protocol in the harbor.cfg
    ${rc}  ${ip}=  Run And Return Rc And Output  ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'
    Log  ${ip}
    ${rc}=  Run And Return Rc  sed "s/^IP=.*/IP=${ip}/g" -i ./tests/generateCerts.sh
    Log  ${rc}
    ${out}=  Run  cat ./tests/generateCerts.sh
    Log  ${out}	
    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/generateCerts.sh
    Should Be Equal As Integers  ${rc}  0

Compile and Up Harbor With Source Code
    [Arguments]  ${golang_image}=golang:${GOLANG_VERSION}  ${clarity_image}=vmware/harbor-clarity-ui-builder:${CLAIR_BUILDER}  ${with_notary}=true  ${with_clair}=true
    ${rc}  ${output}=  Run And Return Rc And Output  docker pull ${clarity_image}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker pull ${golang_image}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  make install GOBUILDIMAGE=${golang_image} COMPILETAG=compile_golangimage CLARITYIMAGE=${clarity_image} NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair} HTTPPROXY=
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    Sleep  20
	
Wait for Harbor Ready
    [Arguments]  ${protocol}  ${HARBOR_IP}
    Log To Console  Waiting for Harbor to Come Up...
    :FOR  ${i}  IN RANGE  20
    \  ${out}=  Run  curl -k ${protocol}://${HARBOR_IP}
    \  Log  ${out}
    \  ${status}=  Run Keyword And Return Status  Should Not Contain  ${out}  502 Bad Gateway
    \  ${status}=  Run Keyword If  ${status}  Run Keyword And Return Status  Should Not Contain  ${out}  Connection refused
    \  ${status}=  Run Keyword If  ${status}  Run Keyword And Return Status  Should Contain  ${out}  <title>Harbor</title>
    \  Return From Keyword If  ${status}  ${HARBOR_IP}
    \  Sleep  30s
    Fail Harbor failed to come up properly!
