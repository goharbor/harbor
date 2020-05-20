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
Resource  ../../resources/Util.robot
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${OIDC_HOSTNAME}
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Get Harbor Version
#Just get harbor version and log it
    Get Harbor Version

Test Case - OIDC User Sign In
    #Sign in with all 9 users is for user population, other test cases might use these users.
    Sign In Harbor With OIDC User    ${HARBOR_URL}
    Sleep  2
    Sign In Harbor With OIDC User    ${HARBOR_URL}    test2
    Sleep  2
    Sign In Harbor With OIDC User    ${HARBOR_URL}    test3
    Sleep  2
    Sign In Harbor With OIDC User    ${HARBOR_URL}    test4
    Sleep  2
    Sign In Harbor With OIDC User    ${HARBOR_URL}    test5
    Sleep  2
    Sign In Harbor With OIDC User    ${HARBOR_URL}    test6
    Sleep  2
    Sign In Harbor With OIDC User    ${HARBOR_URL}    test7
    Sleep  2
    Sign In Harbor With OIDC User    ${HARBOR_URL}    test8
    Sleep  2
    Sign In Harbor With OIDC User    ${HARBOR_URL}    test9
    Sleep  2
    Close Browser

Test Case - Create An New Project
    Sign In Harbor With OIDC User  ${HARBOR_URL}
    ${d}=    Get Current Date    result_format=%m%s
    Create An New Project And Go Into Project  test${d}
    Close Browser

Test Case - Delete A Project
    Init Chrome Driver
    Sign In Harbor With OIDC User  ${HARBOR_URL}
    ${secret}=  Get Secrete By API  ${HARBOR_URL}
    Delete A Project Without Sign In Harbor   harbor_ip=${OIDC_HOSTNAME}  username=${OIDC_USERNAME}  password=${secret}
    Close Browser

Test Case - Manage Project Member
    Init Chrome Driver
    Sign In Harbor With OIDC User  ${HARBOR_URL}
    ${secret}=  Get Secrete By API  ${HARBOR_URL}
    Manage Project Member Without Sign In Harbor  sign_in_user=${OIDC_USERNAME}  sign_in_pwd=${secret}  test_user1=test2  test_user2=test3  is_oidc_mode=${true}
    Close Browser

Test Case - Generate User CLI Secret
    Init Chrome Driver
    ${d}=    Get current Date  result_format=%m%s
    ${image}=  Set Variable  hello-world
    Sign In Harbor With OIDC User  ${HARBOR_URL}
    Create An New Project And Go Into Project  project${d}
    ${secret_old}=  Get Secrete By API  ${HARBOR_URL}
    Push image  ip=${ip}  user=${OIDC_USERNAME}  pwd=${secret_old}  project=project${d}  image=${image}
    ${secret_new}=  Generate And Return Secret  ${HARBOR_URL}
    Log To Console  ${secret_old}
    Log To Console  ${secret_new}
    Should Not Be Equal As Strings  '${secret_old}'  '${secret_new}'
    Cannot Docker Login Harbor  ${ip}  ${OIDC_USERNAME}  ${secret_old}
    Pull image  ${ip}  ${OIDC_USERNAME}  ${secret_new}  project${d}  ${image}
    Push image  ${ip}  ${OIDC_USERNAME}  ${secret_new}  project${d}  ${image}

Test Case - Helm CLI Push
    Init Chrome Driver
    Sign In Harbor With OIDC User  ${HARBOR_URL}
    ${secret}=  Get Secrete By API  ${HARBOR_URL}
    Helm CLI Push Without Sign In Harbor  ${OIDC_USERNAME}  ${secret}