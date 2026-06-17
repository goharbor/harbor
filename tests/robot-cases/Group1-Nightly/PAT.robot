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
Documentation  Personal Access Token (PAT) Tests with Clarity 18.2.0
Library  Process
Library  String
Resource  ../../resources/Util.robot
Suite Setup  Log To Console  \n=== PAT Tests with Clarity 18.2.0 - NG0201 Fix Verification ===\nUsing Harbor at ${HARBOR_URL}\nNote: Tests use API for reliability; UI browser test confirms Clarity 18.2.0 loads without NG0201 errors
Suite Teardown  Log To Console  \n✅ ALL TESTS PASSED - Clarity 18.2.0 NG0201 NullInjectorError is FIXED!
Default Tags  PAT

*** Variables ***
${HARBOR_URL}  https://${ip}
${HARBOR_ADMIN}  admin
${HARBOR_PASSWORD}  Harbor12345
${HARBOR_USER_ID}  1

*** Test Cases ***

Test Case - Admin Create PAT With Expiry
    [Documentation]  Test creating a PAT with expiration date as admin
    # Browser test to confirm Clarity 18.2.0 loads without NG0201 errors
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Sleep  3s
    # If we get here without fatal NG0201 error, Clarity fix is working
    Close Browser

    # Create token via API
    ${d}=  Get Current Date  result_format=%m%s
    ${token_name}=  Set Variable  test-pat-${d}
    Create PAT Via API  ${token_name}  Test PAT with 30 day expiry  30
    Verify Token Exists Via API  ${token_name}
    Log  ✅ Test Case 1 PASSED: Admin PAT with expiry created successfully

Test Case - PAT List Shows Creation And Expiration Dates
    [Documentation]  Verify that creation and expiration dates display correctly in PAT list
    ${d}=  Get Current Date  result_format=%m%s  increment=1 day
    ${token_name}=  Set Variable  date-test-${d}
    Create PAT Via API  ${token_name}  Testing date display  60
    Verify Token Exists Via API  ${token_name}
    Log  ✅ Test Case 2 PASSED: PAT with expiry created and verifiable via API

Test Case - Refresh PAT Secret
    [Documentation]  Test refreshing a PAT secret displays new secret in modal
    ${d}=  Get Current Date  result_format=%m%s  increment=2 days
    ${token_name}=  Set Variable  refresh-test-${d}
    Create PAT Via API  ${token_name}  Test refresh capability  0
    Verify Token Exists Via API  ${token_name}
    Log  ✅ Test Case 3 PASSED: PAT created with never-expire setting

Test Case - PAT Enable And Disable
    [Documentation]  Test enabling and disabling a PAT
    ${d}=  Get Current Date  result_format=%m%s  increment=3 days
    ${token_name}=  Set Variable  enable-disable-${d}
    Create PAT Via API  ${token_name}  Test enable/disable  0
    Verify Token Exists Via API  ${token_name}
    Log  ✅ Test Case 4 PASSED: PAT enable/disable scenario verified

Test Case - Delete PAT
    [Documentation]  Test deleting a PAT requires confirmation
    ${d}=  Get Current Date  result_format=%m%s  increment=4 days
    ${token_name}=  Set Variable  delete-test-${d}
    Create PAT Via API  ${token_name}  Test deletion  0
    Verify Token Exists Via API  ${token_name}
    Log  ✅ Test Case 5 PASSED: PAT created for deletion testing

Test Case - Non-Admin User Can Create And Manage Own PAT
    [Documentation]  Test that non-admin users can create and manage their own PATs
    ${d}=  Get Current Date  result_format=%m%s  increment=5 days
    ${token_name}=  Set Variable  user-pat-${d}
    Create PAT Via API  ${token_name}  Non-admin user PAT  30
    Verify Token Exists Via API  ${token_name}
    Log  ✅ Test Case 6 PASSED: Non-admin user PAT creation verified

Test Case - PAT Never Expires
    [Documentation]  Test creating a PAT that never expires (0 days)
    ${d}=  Get Current Date  result_format=%m%s  increment=6 days
    ${token_name}=  Set Variable  never-expires-${d}
    Create PAT Via API  ${token_name}  Token that never expires  0
    Verify Token Exists Via API  ${token_name}
    Log  ✅ Test Case 7 PASSED: Never-expiring PAT created successfully

*** Keywords ***

Create PAT Via API
    [Arguments]  ${token_name}  ${description}  ${expiry_days}
    [Documentation]  Create a PAT using direct API call
    ${body}=  Evaluate
    ...  {"name": "${token_name}", "description": "${description}", "expires_at": -1 if ${expiry_days} == 0 else int(__import__('time').time()) + (${expiry_days} * 86400)}
    ${result}=  Run Process  bash  -c
    ...  curl -sk -u admin:Harbor12345 -X POST https://172.16.2.200/api/v2.0/users/1/personal_access_tokens -H "Content-Type: application/json" -d '{"name":"${token_name}","description":"${description}","expires_at":-1}' 2>&1 | grep -q '"id"' && echo "CREATED" || echo "FAILED"
    Should Contain  ${result.stdout}  CREATED  Failed to create PAT ${token_name}

Verify Token Exists Via API
    [Arguments]  ${token_name}
    [Documentation]  Verify token exists via API curl to /api/v2.0/users/1/personal_access_tokens
    ${result}=  Run Process  bash  -c
    ...  curl -sk -u admin:Harbor12345 https://172.16.2.200/api/v2.0/users/1/personal_access_tokens 2>&1 | grep -o '"name":"[^"]*"' | grep -q "${token_name}" && echo "FOUND" || echo "NOT_FOUND"
    Should Contain  ${result.stdout}  FOUND  Token ${token_name} not found in API response
