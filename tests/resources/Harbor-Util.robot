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
Library  Selenium2Library
Library  OperatingSystem

*** Variables ***

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
    [Arguments]  ${with_notary}=true  ${with_clair}=true  ${with_chartmuseum}=true
    ${rc}  ${output}=  Run And Return Rc And Output  make start -e NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair} CHARTFLAG=${with_chartmuseum}
    Log  ${rc}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0

Down Harbor
    [Arguments]  ${with_notary}=true  ${with_clair}=true  ${with_chartmuseum}=true
    ${rc}  ${output}=  Run And Return Rc And Output  echo "Y" | make down -e NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair} CHARTFLAG=${with_chartmuseum}
    Log  ${rc}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0

Package Harbor Offline
    [Arguments]  ${with_notary}=true  ${with_clair}=true  ${with_migrator}=true  ${with_chartmuseum}=true
    Log To Console  \nStart Docker Daemon
    Start Docker Daemon Locally
    Log To Console  \n\nmake package_offline VERSIONTAG=%{Harbor_Assets_Version} PKGVERSIONTAG=%{Harbor_Package_Version} UIVERSIONTAG=%{Harbor_UI_Version} NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair} MIGRATORFLAG=${with_migrator} CHARTFLAG=${with_chartmuseum} HTTPPROXY=
    ${rc}  ${output}=  Run And Return Rc And Output  make package_offline VERSIONTAG=%{Harbor_Assets_Version} PKGVERSIONTAG=%{Harbor_Package_Version} UIVERSIONTAG=%{Harbor_UI_Version} NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair} MIGRATORFLAG=${with_migrator} CHARTFLAG=${with_chartmuseum} HTTPPROXY=
    Log  ${rc}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0

Package Harbor Online
    [Arguments]  ${with_notary}=true  ${with_clair}=true  ${with_migrator}=false  ${with_chartmuseum}=true
    Log To Console  \nStart Docker Daemon
    Start Docker Daemon Locally
    Log To Console  \nmake package_online VERSIONTAG=%{Harbor_Assets_Version} PKGVERSIONTAG=%{Harbor_Package_Version} UIVERSIONTAG=%{Harbor_UI_Version} NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair} MIGRATORFLAG=${with_migrator} CHARTFLAG=${with_chartmuseum} HTTPPROXY=
    ${rc}  ${output}=  Run And Return Rc And Output  make package_online VERSIONTAG=%{Harbor_Assets_Version} PKGVERSIONTAG=%{Harbor_Package_Version} UIVERSIONTAG=%{Harbor_UI_Version} NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair} MIGRATORFLAG=${with_migrator} CHARTFLAG=${with_chartmuseum} HTTPPROXY=
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
    Log To Console  ${ip}
    ${rc}=  Run And Return Rc  mkdir -p /etc/docker/certs.d/${ip}/
    Should Be Equal As Integers  ${rc}  0
    Log To Console  ${notaryServerEndpointNoSubDir}
    ${rc}=  Run And Return Rc  mkdir -p ~/.docker/tls/${notaryServerEndpointNoSubDir}/
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  cp ./harbor_ca.crt /etc/docker/certs.d/${ip}/
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  cp ./harbor_ca.crt ~/.docker/tls/${notaryServerEndpointNoSubDir}/
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  ls -la /etc/docker/certs.d/${ip}/
    Log  ${output}
    ${rc}  ${output}=  Run And Return Rc And Output  ls -la ~/.docker/tls/${notaryServerEndpointNoSubDir}/
    Log  ${output}

Prepare
    [Arguments]  ${with_notary}=true  ${with_clair}=true  ${with_chartmuseum}=true
    ${rc}  ${output}=  Run And Return Rc And Output  make prepare -e NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair} CHARTFLAG=${with_chartmuseum}
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
    [Arguments]  ${with_notary}=true  ${with_clair}=true  ${with_chartmuseum}=true
    ${rc}  ${output}=  Run And Return Rc And Output  make install swagger_client NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair} CHARTFLAG=${with_chartmuseum} HTTPPROXY=
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

Get Harbor Version
    ${rc}  ${output}=  Run And Return Rc And Output  curl -k -X GET --header 'Accept: application/json' 'https://${ip}/api/systeminfo'|grep -i harbor_version
    Log To Console  ${output}
