
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
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${ip}
${SSH_USER}  root
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Vulnerability Data Not Ready
#This case must run before vulnerability db ready
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  library  has_image=${false}
    Vulnerability Not Ready Project Hint
    Switch To Vulnerability Page
    Vulnerability Not Ready Config Hint

Test Case - Disable Scan Schedule
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Vulnerability Page
    Disable Scan Schedule
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Vulnerability Page
    Retry Wait Until Page Contains  None
    Close Browser

Test Case - Scan A Tag In The Repo
    Body Of Scan A Tag In The Repo

Test Case - Scan As An Unprivileged User
    Init Chrome Driver
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  hello-world

    Sign In Harbor  ${HARBOR_URL}  user024  Test1@34
    Go Into Project  library
    Go Into Repo  hello-world
    Select Object  latest
    Scan Is Disabled
    Close Browser

Test Case - Scan Image With Empty Vul
    Init Chrome Driver
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  busybox
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  library
    Go Into Repo  busybox
    Scan Repo  latest  Succeed
    Move To Summary Chart
    Wait Until Page Contains  Unknow
    Close Browser

Test Case - Manual Scan All
    Init Chrome Driver
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  redis
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Vulnerability Page
    Trigger Scan Now
    Navigate To Projects
    Go Into Project  library
    Go Into Repo  redis
    Summary Chart Should Display  latest
    Close Browser

Test Case - View Scan Error
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user026  Test1@34
    Create An New Project  project${d}
    Push Image  ${ip}  user026  Test1@34  project${d}  vmware/photon:1.0
    Go Into Project  project${d}
    Go Into Repo  project${d}/vmware/photon
    Scan Repo  1.0  Fail
    View Scan Error Log
    Close Browser

Test Case - Scan Image On Push
    [Tags]  run-once
    Wait Unitl Vul Data Ready  ${HARBOR_URL}  7200  30
    Init Chrome Driver
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  hello-world
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  library
    Goto Project Config
    Enable Scan On Push
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  memcached
    Navigate To Projects
    Go Into Project  library
    Go Into Repo  memcached
    Summary Chart Should Display  latest
    Close Browser

Test Case - View Scan Results
    [Tags]  run-once
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user025  Test1@34
    Create An New Project  project${d}
    Push Image  ${ip}  user025  Test1@34  project${d}  tomcat
    Go Into Project  project${d}
    Go Into Repo  project${d}/tomcat
    Scan Repo  latest  Succeed
    Summary Chart Should Display  latest
    View Repo Scan Details
    Close Browser

Test Case - Project Level Image Serverity Policy
    [Tags]  run-once
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    ${d}=  get current date  result_format=%m%s
    ${sha256}=  Set Variable  68b49a280d2fbe9330c0031970ebb72015e1272dfa25f0ed7557514f9e5ad7b7
    ${image}=  Set Variable  postgres
    Create An New Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}  sha256=${sha256}
    Go Into Project  project${d}
    Go Into Repo  ${image}
    Scan Repo  ${sha256}  Succeed
    Navigate To Projects
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  0
    Cannot pull image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}  tag=${sha256}
    Close Browser

#Important Note: All CVE IDs in CVE Whitelist cases must unique!
Test Case - Verfiy System Level CVE Whitelist
    [Tags]  run-once
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=    Set Variable    redis
    ${sha256}=  Set Variable  9755880356c4ced4ff7745bafe620f0b63dd17747caedba72504ef7bac882089
    ${signin_user}=    Set Variable  user025
    ${signin_pwd}=    Set Variable  Test1@34
    Sign In Harbor    ${HARBOR_URL}    ${signin_user}    ${signin_pwd}
    Create An New Project    project${d}
    Push Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    sha256=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  1
    Cannot Pull image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image}
    Scan Repo  ${sha256}  Succeed
    Logout Harbor
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Switch To Configuration System Setting
    Add Items To System CVE Whitelist    CVE-2019-12904\nCVE-2011-3389\nCVE-2018-12886\nCVE-2019-3843\nCVE-2018-20839\nCVE-2019-5094\nCVE-2019-15847\nCVE-2019-13627
    Cannot Pull image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Add Items To System CVE Whitelist    CVE-2019-3844
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Delete Top Item In System CVE Whitelist  count=6
    Cannot Pull image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Close Browser

Test Case - Verfiy Project Level CVE Whitelist
    [Tags]  run-once
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=    Set Variable    redis
    ${sha256}=  Set Variable  9755880356c4ced4ff7745bafe620f0b63dd17747caedba72504ef7bac882089
    ${signin_user}=    Set Variable  user025
    ${signin_pwd}=    Set Variable  Test1@34
    Sign In Harbor    ${HARBOR_URL}    ${signin_user}    ${signin_pwd}
    Create An New Project    project${d}
    Push Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    sha256=${sha256}
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  1
    Cannot Pull image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image}
    Scan Repo  ${sha256}  Succeed
    Go Into Project  project${d}
    Add Items to Project CVE Whitelist    CVE-2019-12904\nCVE-2011-3389\nCVE-2018-12886\nCVE-2019-3843\nCVE-2018-20839\nCVE-2019-5094\nCVE-2019-15847\nCVE-2019-13627
    Cannot Pull image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Add Items to Project CVE Whitelist    CVE-2019-3844
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Delete Top Item In Project CVE Whitelist
    Cannot Pull image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Close Browser

Test Case - Verfiy Project Level CVE Whitelist By Quick Way of Add System
    [Tags]  run-once
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=    Set Variable    redis
    ${sha256}=  Set Variable  9755880356c4ced4ff7745bafe620f0b63dd17747caedba72504ef7bac882089
    ${signin_user}=    Set Variable  user025
    ${signin_pwd}=    Set Variable  Test1@34
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Switch To Configuration System Setting
    Add Items To System CVE Whitelist    CVE-2019-12904\nCVE-2011-3389\nCVE-2018-12886\nCVE-2019-3843\nCVE-2018-20839\nCVE-2019-3844\nCVE-2019-5094\nCVE-2019-15847\nCVE-2019-13627
    Logout Harbor
    Sign In Harbor    ${HARBOR_URL}    ${signin_user}    ${signin_pwd}
    Create An New Project    project${d}
    Push Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    sha256=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  1
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image}
    Scan Repo  ${sha256}  Succeed
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Go Into Project  project${d}
    Set Project To Project Level CVE Whitelist
    Cannot Pull image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Add System CVE Whitelist to Project CVE Whitelist By Add System Button Click
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Close Browser