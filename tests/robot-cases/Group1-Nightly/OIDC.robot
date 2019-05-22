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
    Sign In Harbor With OIDC User    ${HARBOR_URL}
    ${d}=    Get Current Date    result_format=%M%S
    Create An New Project  test${d}
    Close Browser

Test Case - Delete A Project
    Init Chrome Driver
    Sign In Harbor With OIDC User    ${HARBOR_URL}
    ${json}=  Run Curl And Return Json  curl -s -k -X GET --header 'Accept: application/json' -u '${HARBOR_ADMIN}:${HARBOR_PASSWORD}' 'https://${ip}/api/users/search?username=${OIDC_USERNAME}'
    ${user_info}=    Set Variable    ${json[0]}
    ${user_id}=    Set Variable    ${user_info["user_id"]}
    ${json}=  Run Curl And Return Json   curl -s -k -X GET --header 'Accept: application/json' -u '${HARBOR_ADMIN}:${HARBOR_PASSWORD}' 'https://${ip}/api/users/${user_id}'
    ${secret}=    Set Variable    ${json["oidc_user_meta"]["secret"]}
    Delete A Project Without Sign In Harbor   harbor_ip=${OIDC_HOSTNAME}  username=${OIDC_USERNAME}  password=${secret}
    Close Browser