# Copyright Project Harbor Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License

*** Settings ***
Documentation  Harbor robot account secret validation tests
Resource  ../../resources/Util.robot
Resource  ../../resources/Harbor-Pages/Robot_Account.robot
Resource  ../../resources/Harbor-Pages/Robot_Account_Elements.robot
Resource  ../../resources/Harbor-Pages/Project_Robot_Account.robot
Resource  ../../resources/Harbor-Pages/Project_Robot_Account_Elements.robot

*** Variables ***

*** Test Cases ***
Test Secret Length Validation Too Short
    [Documentation]  Test that secret less than 8 characters is rejected
    Init Browser
    Sign In Harbor
    Switch To System Admin
    Verify System Robot Secret Validation Error  Test1  Secret must be 8-128 characters long
    Sleep  2

Test Secret Length Validation Too Long
    [Documentation]  Test that secret more than 128 characters is rejected
    Init Browser
    Sign In Harbor
    Switch To System Admin
    ${long_secret}=  Evaluate  'A' * 129
    Verify System Robot Secret Validation Error  ${long_secret}  Secret must be 8-128 characters long
    Sleep  2

Test Secret Missing Uppercase
    [Documentation]  Test that secret without uppercase letter is rejected
    Init Browser
    Sign In Harbor
    Switch To System Admin
    Verify System Robot Secret Validation Error  testsecret1  Secret must contain at least 1 uppercase letter
    Sleep  2

Test Secret Missing Lowercase
    [Documentation]  Test that secret without lowercase letter is rejected
    Init Browser
    Sign In Harbor
    Switch To System Admin
    Verify System Robot Secret Validation Error  TESTSECRET1  Secret must contain at least 1 lowercase letter
    Sleep  2

Test Secret Missing Digit
    [Documentation]  Test that secret without digit is rejected
    Init Browser
    Sign In Harbor
    Switch To System Admin
    Verify System Robot Secret Validation Error  TestSecret  Secret must contain at least 1 digit
    Sleep  2

Test Create System Robot Account With Valid User Secret
    [Documentation]  Test creating system robot account with user-provided secret
    Init Browser
    Sign In Harbor
    Switch To System Admin
    ${d}=  Get Current Date  result_format=%m%s
    ${robot_account_name}=  Create A System Robot Account With User Secret  test_user_secret${d}  ValidSecret123  never  description=Testing user-provided secret  cover_all_system_resources=${true}
    Retry Wait Until Page Contains  You provided your own secret
    Sleep  2

Test Create Project Robot Account With Valid User Secret
    [Documentation]  Test creating project robot account with user-provided secret
    Init Browser
    Sign In Harbor
    Switch To System Admin
    Switch To An New Project  test_secret_project
    ${d}=  Get Current Date  result_format=%m%s
    ${resources}=  Create List  all
    ${robot_account_name}  ${permission_count}=  Create A Project Robot Account With User Secret  test_proj_secret${d}  ProjectSecret123  never  description=Testing project secret  resources=${resources}
    Retry Wait Until Page Contains  You provided your own secret
    Sleep  2

Test Secret Confirmation Mismatch
    [Documentation]  Test that mismatched confirmation shows error
    Init Browser
    Sign In Harbor
    Switch To System Admin
    Retry Element Click  ${new_sys_robot_account_btn}
    Retry Wait Element Should Be Disabled  //button[text()='Next']
    Retry Text Input  ${sys_robot_account_name_input}  test_mismatch
    Retry Text Input  ${sys_robot_account_secret_input}  TestSecret123
    Retry Wait Until Page Contains Element  ${sys_robot_account_secret_confirm_input}
    Retry Text Input  ${sys_robot_account_secret_confirm_input}  DifferentSecret123
    Retry Wait Until Page Contains Element  //span[contains(.,'Secrets do not match')]
    Retry Element Click  //button[@aria-label='Close']
    Sleep  2

Test Secret Visibility Toggle
    [Documentation]  Test toggling secret visibility
    Init Browser
    Sign In Harbor
    Switch To System Admin
    Retry Element Click  ${new_sys_robot_account_btn}
    Retry Text Input  ${sys_robot_account_name_input}  test_visibility
    Retry Text Input  ${sys_robot_account_secret_input}  TestSecret123
    ${initial_type}=  Get Element Attribute  ${sys_robot_account_secret_input}  type
    Should Be Equal  ${initial_type}  password
    Toggle System Robot Secret Visibility
    ${visible_type}=  Get Element Attribute  ${sys_robot_account_secret_input}  type
    Should Be Equal  ${visible_type}  text
    Retry Element Click  //button[@aria-label='Close']
    Sleep  2

Test Secret Edge Case Min Length
    [Documentation]  Test creating account with minimum valid secret length (8 chars)
    Init Browser
    Sign In Harbor
    Switch To System Admin
    ${d}=  Get Current Date  result_format=%m%s
    ${robot_account_name}=  Create A System Robot Account With User Secret  test_min_len${d}  Abcd1234  never  cover_all_system_resources=${true}
    Retry Wait Until Page Contains  You provided your own secret
    Sleep  2

Test Secret Edge Case Max Length
    [Documentation]  Test creating account with maximum valid secret length (128 chars)
    Init Browser
    Sign In Harbor
    Switch To System Admin
    ${d}=  Get Current Date  result_format=%m%s
    ${max_secret}=  Evaluate  'A' + 'a' * 126 + '1'
    ${robot_account_name}=  Create A System Robot Account With User Secret  test_max_len${d}  ${max_secret}  never  cover_all_system_resources=${true}
    Retry Wait Until Page Contains  You provided your own secret
    Sleep  2

Test Secret With Special Characters
    [Documentation]  Test creating account with special characters in secret
    Init Browser
    Sign In Harbor
    Switch To System Admin
    ${d}=  Get Current Date  result_format=%m%s
    ${robot_account_name}=  Create A System Robot Account With User Secret  test_special${d}  Test@Secret#123  never  cover_all_system_resources=${true}
    Retry Wait Until Page Contains  You provided your own secret
    Sleep  2

Test Auto Generated Secret Still Works
    [Documentation]  Test that leaving secret empty still generates secret automatically
    Init Browser
    Sign In Harbor
    Switch To System Admin
    ${d}=  Get Current Date  result_format=%m%s
    ${robot_account_name}  ${token}=  Create A System Robot Account  auto_gen${d}  never  cover_all_system_resources=${true}
    Retry Wait Until Page Contains  Copy secret
    Sleep  2
