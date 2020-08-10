// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

*** Settings ***
Documentation  Harbor BATs
Library  Selenium2Library
Library  OperatingSystem
Library  Process
Default Tags  Bundle

*** Keywords ***
Start Docker Daemon Locally
    ${pid}=  Run  pidof dockerd
    #${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group0-Util/docker_config.sh
    #Log  ${output}
    #Should Be Equal As Integers  ${rc}  0
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

Package Harbor Offline
    [Arguments]  ${with_notary}=true  ${with_clair}=true  ${with_chartmuseum}=true  ${with_trivy}=true
    Log To Console  \nStart Docker Daemon
    Start Docker Daemon Locally
    Log To Console  \nMake Offline Package
    Log To Console  \n\nmake package_offline GOBUILDTAGS="include_oss include_gcs" BASEIMAGETAG=%{Harbor_Build_Base_Tag} NPM_REGISTRY=%{NPM_REGISTRY} VERSIONTAG=%{Harbor_Assets_Version} PKGVERSIONTAG=%{Harbor_Package_Version} NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair} CHARTFLAG=${with_chartmuseum} TRIVYFLAG=${with_trivy} HTTPPROXY=
    ${rc}  ${output}=  Run And Return Rc And Output  make package_offline GOBUILDTAGS="include_oss include_gcs" BASEIMAGETAG=%{Harbor_Build_Base_Tag} NPM_REGISTRY=%{NPM_REGISTRY} VERSIONTAG=%{Harbor_Assets_Version} PKGVERSIONTAG=%{Harbor_Package_Version} NOTARYFLAG=${with_notary} CLAIRFLAG=${with_clair} CHARTFLAG=${with_chartmuseum} TRIVYFLAG=${with_trivy} HTTPPROXY=
    Log To Console  ${rc}
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0

*** Test Cases ***
Distro Harbor Offline
    Package Harbor Offline
