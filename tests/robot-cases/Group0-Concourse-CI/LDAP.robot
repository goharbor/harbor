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
Default Tags  BAT

*** Variables ***
${HARBOR_URL}  https://${ip}

*** Test Cases ***
Test Case - Ldap Verify Cert
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Switch To Configure
    Test Ldap Connection
    Close Browser

Test Case - Ldap Sign in and out
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Switch To Configure
    Init LDAP
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  mike  zhu88jie
    Close Browser

Test Case - Ldap User Create Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  mike  zhu88jie
    Create An New Project  project${d}
    Close Browser

Test Case - Ldap User Push An Image
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  mike  zhu88jie
    Create An New Project  project${d}
    
    Push Image  ${ip}  mike  zhu88jie  project${d}  hello-world:latest
    Go Into Project  project${d}
    Wait Until Page Contains  project${d}/hello-world
    Close Browser

Test Case - Ldap User Can Not login
    Docker Login Fail  ${ip}  test  123456