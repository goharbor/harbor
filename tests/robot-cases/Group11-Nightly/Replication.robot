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
Suite Setup  Nightly Test Setup  ${ip}  ${SSH_PWD}  ${HARBOR_PASSWORD}  ${ip1}
Suite Teardown  Collect Nightly Logs  ${ip}  ${SSH_PWD}  ${ip1}
Default Tags  Replication

*** Variables ***
${HARBOR_URL}  https://${ip}
${SSH_USER}  root
${HARBOR_ADMIN}  admin
*** Test Cases ***
Test Case - Get Harbor Version
#Just get harbor version and log it
    Get Harbor Version

Test Case - Pro Replication Rules Add
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Check New Rule UI Without Endpoint
    Close Browser

Test Case - Endpoint Verification
#This case need vailid info and selfsign cert
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Registries
    Create A New Endpoint  edp1${d}  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  N
    Endpoint Is Pingable
    Enable Certificate Verification
    Endpoint Is Unpingable
    Close Browser

Test Case - Endpoint Add
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Registries
    Create A New Endpoint  testabc  https://${d}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Wait Until Page Contains  testabc
    Close Browser

Test Case - Endpoint Edit
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Registries
    Rename Endpoint  testabc  deletea
    Wait Until Page Contains  deletea
    Close Browser

Test Case - Endpoint Delete  
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Registries
    Delete Endpoint  deletea
    Delete Success  deletea
    Close Browser
   
Test Case - Rule Edit
    Init Chrome Driver
    ${d}=    Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}
    Switch To Registries
    Create A New Endpoint  e${d}  https://ip  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule${d}  project${d}  e${d}  Immediate
    Rename Rule  rule${d}  newname
    Wait Until Page Contains  newname
    Close Browser

Test Case - Rule Delete
    Init Chrome Driver
    ${d}=    Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Delete Rule  newname
    Delete Success  newname
    Close Browser


Test Case - Trigger Immediate
    Init Chrome Driver
    ${d}=    Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}
    Switch To Registries
    Create A New Endpoint  edp${d}  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule${d}  project${d}  edp${d}  Immediate
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world
    Logout Harbor
    #logout and login target
    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  project${d}
    Page Should Contain  hello-world
    Go Into Repo  hello-world
    Page Should Contain  latest
    Close Browser

Test Case - Trigger Manual
    Init Chrome Driver
    ${d}=    Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}
    #using existing endpoint added before for only one replication endpoint
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule${d}  project${d}  edp  Manual
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world
    Trigger Replication Manual  rule${d}
    Logout Harbor
    #logout and login target
    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  project${d}
    Page Should Contain  hello-world
    Go Into Repo  hello-world
    Page Should Contain  latest
    Close Browser

Test Case - Large Image Replicate
    Init Chrome Driver
    ${d}=    Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}
    Push Image with tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ubuntu  16.04  16.04
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule${d}  project${d}  edp  Immediate
    Logout Harbor
    #logout and login target
    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  project${d}
    Page Should Contain  ubuntu
    Go Into Repo  ubuntu
    Page Should Contain  16.04
    Close Browser

Test Case - Proj Replication Jobs Log View
    Init Chrome Driver
    ${d}=    Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  busybox
    Switch To Registries
    Create A New Endpoint  edp${d}  aaa  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule${d}  project${d}  edp${d}  Immediate
    Filter Rule  rule${d}
    Select Rule  rule${d}
    Wait Until Page Contains  transfer 
    Wait Until Page Contains  error
    View Job Log  busybox
    Close Browser

Test Case - Project LeveL Replication Operation
    Init Chrome Driver
    ${d} =  Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  proj${d}
    Go Into Project  proj${d}
    Switch To Replication
    Project Create A Rule With Existing Endpoint  rule${d}  proj${d}  edp  Manual
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  proj${d}  hello-world
    Trigger Replication Manual  rule${d}
    Logout Harbor
    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  proj${d}
    Page Should Contain  hello-world
    Go Into Repo  hello-world
    Page Should Contain  latest
    Close Browser
