
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
Test Case - Sign With Admin
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Close Browser

Test Case - Vulnerability Data Not Ready
#This case must run before vulnerability db ready
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  library  has_image=${false}
    Vulnerability Not Ready Project Hint
    Switch To Vulnerability Page
    Vulnerability Not Ready Config Hint

Test Case - Garbage Collection
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world
    Sleep  2
    Go Into Project  project${d}
    Delete Repo  project${d}

    Switch To Garbage Collection
    Click GC Now
    Logout Harbor
    Sleep  2
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Garbage Collection
    Sleep  1
    Switch To GC History
    Retry Wait Until Page Contains  Finished

    ${rc}  ${output}=  Run And Return Rc And Output  curl -u ${HARBOR_ADMIN}:${HARBOR_PASSWORD} -i --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/system/gc/1/log"
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  3 blobs and 0 manifests eligible for deletion
    #Should Contain  ${output}  Deleting blob:
    Should Contain  ${output}  success to run gc in job.

    Close Browser

Test Case - Create An New Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  test${d}
    Close Browser

Test Case - Delete A Project
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Delete A Project Without Sign In Harbor
    Close Browser

Test Case - Read Only Mode
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}

    Enable Read Only
    Cannot Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  busybox:latest

    Disable Read Only
    Sleep  5
    Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  busybox:latest
    Close Browser

Test Case - Repo Size
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  alpine  2.6  2.6
    Go Into Project  library
    Go Into Repo  alpine
    Wait Until Page Contains  1.92MB
    Close Browser

Test Case - Staticsinfo
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    ${element}=  Set Variable  ${project_statistics_private_repository_icon}
    Wait Until Element Is Visible  ${element}
    ${privaterepocount1}=  Get Statics Private Repo
    ${privateprojcount1}=  Get Statics Private Project
    ${publicrepocount1}=  Get Statics Public Repo
    ${publicprojcount1}=  Get Statics Public Project
    ${totalrepocount1}=  Get Statics Total Repo
    ${totalprojcount1}=  Get Statics Total Project
    Create An New Project  private${d}
    Create An New Project  public${d}  true
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  private${d}  hello-world
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  public${d}  hello-world
    Reload Page
    ${privateprojcount}=  evaluate  ${privateprojcount1}+1
    ${privaterepocount}=  evaluate  ${privaterepocount1}+1
    ${publicprojcount}=  evaluate  ${publicprojcount1}+1
    ${publicrepocount}=  evaluate  ${publicrepocount1}+1
    ${totalrepocount}=  evaluate  ${totalrepocount1}+2
    ${totalprojcount}=  evaluate  ${totalprojcount1}+2
    Navigate To Projects
    Wait Until Element Is Visible  ${element}
    ${privaterepocountStr}=  Convert To String  ${privaterepocount}
    Wait Until Element Contains  ${element}  ${privaterepocountStr}
    ${privaterepocount2}=  Get Statics Private Repo
    ${privateprojcount2}=  get statics private project
    ${publicrepocount2}=  get statics public repo
    ${publicprojcount2}=  get statics public project
    ${totalrepocount2}=  get statics total repo
    ${totalprojcount2}=  get statics total project
    Should Be Equal As Integers  ${privateprojcount2}  ${privateprojcount}
    Should be equal as integers  ${privaterepocount2}  ${privaterepocount}
    Should be equal as integers  ${publicprojcount2}  ${publicprojcount}
    Should be equal as integers  ${publicrepocount2}  ${publicrepocount}
    Should be equal as integers  ${totalprojcount2}  ${totalprojcount}
    Should be equal as integers  ${totalrepocount2}  ${totalrepocount}

Test Case - Push Image
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  test${d}

    Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  test${d}  hello-world:latest
    Go Into Project  test${d}
    Wait Until Page Contains  test${d}/hello-world
