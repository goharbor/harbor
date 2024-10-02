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

Test Case - Update OIDC Provider Name
    [Tags]  oidc_provider_name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  is_oidc=${true}
    # Set OIDC Provider Name to TestDex
    Switch To Configuration Authentication
    Retry Text Input  //input[@id='oidcName']  TestDex
    Retry Element Click  ${config_auth_save_button_xpath}
    Logout Harbor
    Retry Wait Until Page Contains Element  //span[normalize-space()='LOGIN WITH TestDex']
    Close Browser

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
    Push image  ${ip}  ${OIDC_USERNAME}  ${secret_old}  project${d}  ${image}
    ${secret_new}=  Generate And Return Secret  ${HARBOR_URL}
    Log To Console  ${secret_old}
    Log To Console  ${secret_new}
    Should Not Be Equal As Strings  '${secret_old}'  '${secret_new}'
    Cannot Docker Login Harbor  ${ip}  ${OIDC_USERNAME}  ${secret_old}
    Pull image  ${ip}  ${OIDC_USERNAME}  ${secret_new}  project${d}  ${image}
    Push image  ${ip}  ${OIDC_USERNAME}  ${secret_new}  project${d}  ${image}
    Close Browser

Test Case - Onboard OIDC User Sign In
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  is_oidc=${true}
    Check Automatic Onboarding And Save
    Logout Harbor
    Sign In Harbor With OIDC User  ${HARBOR_URL}  test8  is_onboard=${true}
    Logout Harbor
	Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  is_oidc=${true}
    Set User Name Claim And Save  email
    Logout Harbor
    Sign In Harbor With OIDC User  ${HARBOR_URL}  test9  is_onboard=${true}  username_claim=email
    Logout Harbor
	Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  is_oidc=${true}
    Set User Name Claim And Save  ${null}
    Sleep  2
    Close Browser

Test Case - OIDC Group User
    Init Chrome Driver
    ${d}=    Get current Date  result_format=%m%s
    ${image}=  Set Variable  hello-world
    ${admin_user}=  Set Variable  admin_user
    ${admin_pwd}=  Set Variable  zhu88jie
    ${user}=  Set Variable  mike
    ${pwd}=  Set Variable  ${admin_pwd}
    Sign In Harbor With OIDC User  ${HARBOR_URL}  username=${admin_user}  password=${admin_pwd}  login_with_provider=ldap
    Switch To Registries
    Create A New Endpoint    harbor    test_oidc_admin    https://${LOCAL_REGISTRY}    ${null}    ${null}    Y
    ${secret}=  Get Secrete By API  ${HARBOR_URL}  username=${admin_user}
    Push image  ${ip}  ${admin_user}  ${secret}  library  ${image}
    Logout Harbor
    Sign In Harbor With OIDC User  ${HARBOR_URL}  username=${user}  password=${pwd}  login_with_provider=ldap
    ${output}=  Run Keyword And Ignore Error  Switch To Configure
    Should Be Equal As Strings  '${output[0]}'  'FAIL'
    Close Browser

Test Case - Delete An OIDC User In Local DB
    Init Chrome Driver
    # sign in with admin role
    ${admin_user}=  Set Variable  admin_user
    ${admin_pwd}=  Set Variable  zhu88jie
    Sign In Harbor With OIDC User  ${HARBOR_URL}  username=${admin_user}  password=${admin_pwd}  login_with_provider=ldap
    # shoule be able to delete an OIDC user
    Able To Delete An OIDC User
    # Re-sign in with the deleted user, will get it back
    Sign In Harbor With OIDC User    ${HARBOR_URL}    test7
    Sleep  2
    Sign In Harbor With OIDC User  ${HARBOR_URL}  username=${admin_user}  password=${admin_pwd}  login_with_provider=ldap
    Should Contain Target User
    Close Browser

Test Case - OIDC Group Filter
    [Tags]  group_filter
    Init Chrome Driver
    ${oidc_user}=  Set Variable  mike02
    ${oidc_pwd}=  Set Variable  zhu88jie
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  is_oidc=${true}
    Retry Element Click  //clr-vertical-nav//span[contains(.,'Groups')]
    Retry Wait Until Page Contains Element  //clr-dg-pagination//div[contains(@class, 'pagination-description')]
    ${total}=  Get Text  //clr-dg-pagination//div[contains(@class, 'pagination-description')]
    # Delete all groups
    Run Keyword If  '${total}' != '0 items'  Run Keywords  Retry Element Click  //div[@class='clr-checkbox-wrapper']//label[contains(@class,'clr-control-label')]  AND  Retry Button Click  //button[contains(.,'Delete')]  AND  Retry Button Click  //button[contains(.,'DELETE')]
    # Set OIDCGroupFilter to .*users
    Switch To Configuration Authentication
    Retry Text Input  //*[@id='OIDCGroupFilter']  .*users
    Retry Element Click  ${config_auth_save_button_xpath}
    Logout Harbor
    # Login to the Harbor using OIDC user
    Sign In Harbor With OIDC User  ${HARBOR_URL}  username=${oidc_user}  password=${oidc_pwd}  login_with_provider=ldap
    Logout Harbor
    # Check that there is only one harbor_users group
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  is_oidc=${true}
    Retry Element Click  //clr-vertical-nav//span[contains(.,'Groups')]
    Retry Wait Until Page Contains Element  //app-group//clr-dg-row//clr-dg-cell[text()='harbor_users']
    ${count}=  Get Element Count  //app-group//clr-dg-row
    Should Be Equal As Integers  ${count}  1
    # Reset OIDCGroupFilter
    Switch To Configuration Authentication
    Clear Field Of Characters  //*[@id='OIDCGroupFilter']  7
    Retry Element Click  ${config_auth_save_button_xpath}
    Logout Harbor
    # Login to the Harbor using OIDC user
    Sign In Harbor With OIDC User  ${HARBOR_URL}  username=${oidc_user}  password=${oidc_pwd}  login_with_provider=ldap
    Logout Harbor
    # Check that there are more than one groups
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  is_oidc=${true}
    Retry Element Click  //clr-vertical-nav//span[contains(.,'Groups')]
    Retry Wait Until Page Contains Element  //app-group//clr-dg-row//clr-dg-cell[text()='harbor_users']
    ${count}=  Get Element Count  //app-group//clr-dg-row
    Should Be True  ${count} > 1
    Close Browser
