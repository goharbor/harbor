// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
Documentation  Admin View Project
Resource  ../../resources/Util.robot
Default Tags  regression

*** Test Cases ***
Test Case - Admin View Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Create An New Project  test${d}1
    Create An New Public Project  test${d}2
    Close Browser

    ${rc}  ${ip}=  Run And Return Rc And Output  ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'
    Log To Console  ${ip}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  docker pull hello-world
    Log  ${rc}
    ${rc}=  Run And Return Rc  docker pull busybox
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker login -u tester${d} -p Test1@34 ${ip}
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  docker tag hello-world ${ip}/test${d}1/hello-world:latest
    Log  ${rc}
    ${rc}=  Run And Return Rc  docker tag hello-world ${ip}/test${d}2/busybox:latest
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  docker push ${ip}/test${d}1/hello-world:latest
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  docker push ${ip}/test${d}2/busybox:latest
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0

    Init Chrome Driver
    Go To    http://localhost
    Sleep  2
    ${title}=  Get Title
    Should Be Equal  ${title}  Harbor
    Sign In Harbor  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Sleep  2
    Wait Until Page Contains  test${d}1
    Wait Until Page Contains  test${d}2
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/nav/section/a[2]
    Sleep  2
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/nav/section/a[1]
    Sleep  2
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/project/div/div/list-project/clr-datagrid/div/div/div[2]/clr-dg-row[1]/clr-dg-row-master/clr-dg-cell[1]/a
    Sleep
    Wait Until Page Contains  test${d}1/hello-world
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/project-detail/nav/ul/li[2]/a
    Sleep  2
    Wait Until Page Contains  tester${d}
