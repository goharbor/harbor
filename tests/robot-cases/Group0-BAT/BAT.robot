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
Suite Setup  Install Harbor to Test Server
Suite Teardown  Down Harbor
Default Tags  BAT

*** Variables ***
${HARBOR_URL}  https://${ip}

*** Test Cases ***
Test Case - Registry Basic Verfication
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Create An New Project With New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@harbortest.com  realname=tester${d}  newPassword=Test1@34  comment=harbor  projectname=project${d}  public=true
    Push image  ${ip}  tester${d}  Test1@34  project${d}  busybox:latest
    Pull image  ${ip}  tester${d}  Test1@34  project${d}  busybox:latest
    Go Into Project  project${d}
    Delete Repo  project${d}  busybox

    Close Browser

Test Case - Ldap Basic Verfication
    Switch To LDAP
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Switch To Configure
    Init LDAP
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  mike  zhu88jie
    Close Browser

Test Case - Run LDAP Group Related API Test
    Harbor API Test  ./tests/apitests/python/test_ldap_admin_role.py
    Harbor API Test  ./tests/apitests/python/test_user_group.py
    Harbor API Test  ./tests/apitests/python/test_assign_role_to_ldap_group.py
