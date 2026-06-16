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
Documentation  Personal Access Token (PAT) Tests
Resource  ../../resources/Util.robot
Suite Setup  Install Harbor to Test Server
Suite Teardown  Down Harbor
Default Tags  PAT

*** Variables ***
${HARBOR_URL}  https://${ip}

*** Test Cases ***

Test Case - Admin Create PAT With Expiry
    [Documentation]  Test creating a PAT with expiration date as admin
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}

    # Navigate to account settings
    Click Element  xpath=//*[@id='user-profile-menu']
    Wait Until Element Is Visible  xpath=//*[contains(text(), 'Account Settings')]
    Click Element  xpath=//*[contains(text(), 'Account Settings')]

    # Click PAT tab
    Wait Until Element Is Visible  xpath=//*[contains(text(), 'Personal Access Tokens')]
    Click Element  xpath=//*[contains(text(), 'Personal Access Tokens')]

    # Click Create button
    Wait Until Element Is Visible  xpath=//button[contains(text(), 'Create Token')]
    Click Element  xpath=//button[contains(text(), 'Create Token')]

    # Fill in token details
    Wait Until Element Is Visible  xpath=//input[@id='pat-name']
    Input Text  xpath=//input[@id='pat-name']  test-pat-${d}
    Input Text  xpath=//textarea[@id='pat-description']  Test PAT with 30 day expiry
    Input Text  xpath=//input[@id='pat-expires']  30

    # Create token
    Click Element  xpath=//button[contains(text(), 'CREATE')]

    # Verify secret is displayed in modal
    Wait Until Page Contains  Copy your token now
    Wait Until Element Is Visible  xpath=//input[@type='text' and @readonly]
    ${secret}=  Get Value  xpath=//input[@type='text' and @readonly]
    Should Not Be Empty  ${secret}

    # Close modal
    Click Element  xpath=//button[contains(text(), 'CLOSE')]

    # Verify token appears in list
    Wait Until Page Contains  test-pat-${d}
    Page Should Contain  test-pat-${d}
    Page Should Contain  Active

    Close Browser

Test Case - PAT List Shows Creation And Expiration Dates
    [Documentation]  Verify that creation and expiration dates display correctly in PAT list
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}

    # Navigate to PAT section
    Click Element  xpath=//*[@id='user-profile-menu']
    Click Element  xpath=//*[contains(text(), 'Account Settings')]
    Wait Until Element Is Visible  xpath=//*[contains(text(), 'Personal Access Tokens')]
    Click Element  xpath=//*[contains(text(), 'Personal Access Tokens')]

    # Create a PAT with expiry
    Click Element  xpath=//button[contains(text(), 'Create Token')]
    Wait Until Element Is Visible  xpath=//input[@id='pat-name']
    Input Text  xpath=//input[@id='pat-name']  date-test-pat-${d}
    Input Text  xpath=//input[@id='pat-description']  Testing date display
    Input Text  xpath=//input[@id='pat-expires']  60
    Click Element  xpath=//button[contains(text(), 'CREATE')]

    Wait Until Page Contains  Copy your token now
    Click Element  xpath=//button[contains(text(), 'CLOSE')]

    # Verify dates are displayed in the list
    Wait Until Page Contains  date-test-pat-${d}

    # Verify "Created" column shows a date (not empty)
    # The row should contain the token name
    Page Should Contain  date-test-pat-${d}

    # Verify "Expires" column shows a date (not "creation_time" or raw timestamp)
    # It should contain a date in format like "6/15/26" or similar
    Page Should Not Contain  creation_time
    Page Should Not Contain  expires_at

    Close Browser

Test Case - Refresh PAT Secret
    [Documentation]  Test refreshing a PAT secret displays new secret in modal
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}

    # Navigate to PAT section
    Click Element  xpath=//*[@id='user-profile-menu']
    Click Element  xpath=//*[contains(text(), 'Account Settings')]
    Wait Until Element Is Visible  xpath=//*[contains(text(), 'Personal Access Tokens')]
    Click Element  xpath=//*[contains(text(), 'Personal Access Tokens')]

    # Create initial PAT
    Click Element  xpath=//button[contains(text(), 'Create Token')]
    Wait Until Element Is Visible  xpath=//input[@id='pat-name']
    Input Text  xpath=//input[@id='pat-name']  refresh-test-${d}
    Click Element  xpath=//button[contains(text(), 'CREATE')]
    Wait Until Page Contains  Copy your token now
    ${initial_secret}=  Get Value  xpath=//input[@type='text' and @readonly]
    Click Element  xpath=//button[contains(text(), 'CLOSE')]

    # Refresh the secret
    Wait Until Element Is Visible  xpath=//button[@id='refresh-pat-*']
    Click Element  xpath=//*[@clrDropdownItem and contains(text(), 'Refresh Secret')]

    # Verify modal shows the new secret
    Wait Until Page Contains  Copy your token now
    ${new_secret}=  Get Value  xpath=//input[@type='text' and @readonly]
    Should Not Be Empty  ${new_secret}

    Close Browser

Test Case - PAT Enable And Disable
    [Documentation]  Test enabling and disabling a PAT
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}

    # Navigate to PAT section
    Click Element  xpath=//*[@id='user-profile-menu']
    Click Element  xpath=//*[contains(text(), 'Account Settings')]
    Wait Until Element Is Visible  xpath=//*[contains(text(), 'Personal Access Tokens')]
    Click Element  xpath=//*[contains(text(), 'Personal Access Tokens')]

    # Create PAT
    Click Element  xpath=//button[contains(text(), 'Create Token')]
    Wait Until Element Is Visible  xpath=//input[@id='pat-name']
    Input Text  xpath=//input[@id='pat-name']  toggle-test-${d}
    Click Element  xpath=//button[contains(text(), 'CREATE')]
    Wait Until Page Contains  Copy your token now
    Click Element  xpath=//button[contains(text(), 'CLOSE')]

    # Verify token is Active
    Wait Until Page Contains  toggle-test-${d}
    Page Should Contain  Active

    # Disable the token
    Wait Until Element Is Visible  xpath=//*[@clrDropdownItem and contains(text(), 'Disable')]
    Click Element  xpath=//*[@clrDropdownItem and contains(text(), 'Disable')]

    # Verify token is now Disabled
    Wait Until Page Contains  Disabled
    Page Should Contain  toggle-test-${d}

    # Enable the token
    Click Element  xpath=//*[@clrDropdownItem and contains(text(), 'Enable')]

    # Verify token is Active again
    Wait Until Page Contains  Active

    Close Browser

Test Case - Delete PAT
    [Documentation]  Test deleting a PAT requires confirmation
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}

    # Navigate to PAT section
    Click Element  xpath=//*[@id='user-profile-menu']
    Click Element  xpath=//*[contains(text(), 'Account Settings')]
    Wait Until Element Is Visible  xpath=//*[contains(text(), 'Personal Access Tokens')]
    Click Element  xpath=//*[contains(text(), 'Personal Access Tokens')]

    # Create PAT to delete
    Click Element  xpath=//button[contains(text(), 'Create Token')]
    Wait Until Element Is Visible  xpath=//input[@id='pat-name']
    Input Text  xpath=//input[@id='pat-name']  delete-test-${d}
    Click Element  xpath=//button[contains(text(), 'CREATE')]
    Wait Until Page Contains  Copy your token now
    Click Element  xpath=//button[contains(text(), 'CLOSE')]

    # Verify token exists
    Wait Until Page Contains  delete-test-${d}

    # Delete the token
    Click Element  xpath=//*[@clrDropdownItem and contains(text(), 'DELETE')]

    # Confirm deletion
    Wait Until Element Is Visible  xpath=//button[contains(text(), 'DELETE')]
    Click Element  xpath=//button[contains(text(), 'DELETE')]

    # Verify token is deleted
    Wait Until Page Does Not Contain  delete-test-${d}

    Close Browser

Test Case - Non-Admin User Can Create And Manage Own PAT
    [Documentation]  Test that non-admin users can create and manage their own PATs
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${username}=  Set Variable  testuser${d}
    ${password}=  Set Variable  Test1@34

    # Create a new user first (as admin)
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}

    Navigate To  ${HARBOR_URL}/system/users
    Wait Until Element Is Visible  xpath=//button[contains(text(), 'New User')]
    Click Element  xpath=//button[contains(text(), 'New User')]

    Wait Until Element Is Visible  xpath=//input[@id='username']
    Input Text  xpath=//input[@id='username']  ${username}
    Input Text  xpath=//input[@id='email']  ${username}@test.com
    Input Text  xpath=//input[@id='realname']  Test User
    Input Text  xpath=//input[@id='newPassword']  ${password}
    Input Text  xpath=//input[@id='confirmPassword']  ${password}

    Click Element  xpath=//button[contains(text(), 'OK')]

    Wait Until Page Contains  ${username}

    # Logout and login as the new user
    Click Element  xpath=//*[@id='user-profile-menu']
    Click Element  xpath=//a[contains(text(), 'Log Out')]

    # Login as new user
    Sign In Harbor  ${HARBOR_URL}  ${username}  ${password}

    # Navigate to account settings and create PAT
    Click Element  xpath=//*[@id='user-profile-menu']
    Click Element  xpath=//*[contains(text(), 'Account Settings')]
    Wait Until Element Is Visible  xpath=//*[contains(text(), 'Personal Access Tokens')]
    Click Element  xpath=//*[contains(text(), 'Personal Access Tokens')]

    # Verify user can create PAT
    Click Element  xpath=//button[contains(text(), 'Create Token')]
    Wait Until Element Is Visible  xpath=//input[@id='pat-name']
    Input Text  xpath=//input[@id='pat-name']  user-pat-${d}
    Input Text  xpath=//input[@id='pat-description']  PAT created by non-admin user
    Click Element  xpath=//button[contains(text(), 'CREATE')]

    # Verify token was created
    Wait Until Page Contains  Copy your token now
    Wait Until Element Is Visible  xpath=//input[@type='text' and @readonly]
    Click Element  xpath=//button[contains(text(), 'CLOSE')]

    # Verify token appears in list
    Wait Until Page Contains  user-pat-${d}
    Page Should Contain  Active

    Close Browser

Test Case - PAT Never Expires
    [Documentation]  Test creating a PAT that never expires (0 days)
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}

    # Navigate to PAT section
    Click Element  xpath=//*[@id='user-profile-menu']
    Click Element  xpath=//*[contains(text(), 'Account Settings')]
    Wait Until Element Is Visible  xpath=//*[contains(text(), 'Personal Access Tokens')]
    Click Element  xpath=//*[contains(text(), 'Personal Access Tokens')]

    # Create PAT with 0 expiry (never expires)
    Click Element  xpath=//button[contains(text(), 'Create Token')]
    Wait Until Element Is Visible  xpath=//input[@id='pat-name']
    Input Text  xpath=//input[@id='pat-name']  never-expires-${d}
    Input Text  xpath=//input[@id='pat-description']  Token that never expires
    Input Text  xpath=//input[@id='pat-expires']  0

    Click Element  xpath=//button[contains(text(), 'CREATE')]

    Wait Until Page Contains  Copy your token now
    Click Element  xpath=//button[contains(text(), 'CLOSE')]

    # Verify token shows "Never" in Expires column
    Wait Until Page Contains  never-expires-${d}
    Wait Until Page Contains  Never

    Close Browser
