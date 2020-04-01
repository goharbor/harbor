
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
Test Case - Get Harbor Version
#Just get harbor version and log it
    Get Harbor Version

Test Case - Trivy Is Default Scanner And It Is Immutable
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Scanners Page
    Should Display The Default Trivy Scanner
    Trivy Is Immutable Scanner

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
    Body Of Scan A Tag In The Repo Use Trivy

Test Case - Scan As An Unprivileged User
    Init Chrome Driver
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  hello-world

    Sign In Harbor  ${HARBOR_URL}  user024  Test1@34
    Go Into Project  library
    Go Into Repo  hello-world
    Select Object  latest
    Scan Is Disabled
    Close Browser
# chose a emptyVul repo
Test Case - Scan Image With Empty Vul
    Init Chrome Driver
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  photon:2.0_scan
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  library
    Go Into Repo  library/photon
    Scan Repo  2.0  Succeed
    Move To Summary Chart
    Wait Until Page Contains  No vulnerability
    Close Browser
Test Case - Manual Scan All
    Init Chrome Driver
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  redis
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Vulnerability Page
    Trigger Scan Now And Wait Until The Result Appears
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
    Push Image  ${ip}  user026  Test1@34  project${d}  busybox:latest
    Go Into Project  project${d}
    Go Into Repo  project${d}/busybox
    Scan Repo  latest  Fail
    View Scan Error Log
    Close Browser

Test Case - Scan Image On Push
    [Tags]  run-once
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
