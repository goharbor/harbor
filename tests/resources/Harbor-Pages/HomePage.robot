# Copyright Project Harbor Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License

*** Settings ***
Documentation  This resource provides any keywords related to the Harbor private registry appliance
Resource  ../../resources/Util.robot

*** Variables ***

*** Keywords ***
Sign In Harbor
    [Arguments]  ${url}  ${user}  ${pw}
    Go To    ${url}
    Retry Wait Element  ${harbor_span_title}
    Retry Wait Element  ${login_name}
    Retry Wait Element  ${login_pwd}
    Input Text  ${login_name}  ${user}
    Input Text  ${login_pwd}  ${pw}
    Retry Wait Element  ${login_btn}
    Retry Button Click  ${login_btn}
    Log To Console  ${user}
    Retry Wait Element  xpath=//span[contains(., '${user}')]

Capture Screenshot And Source
    Capture Page Screenshot
    Log Source

Sign Up Should Not Display
    Retry Wait Until Page Not Contains Element  xpath=${sign_up_button_xpath}

Create An New User
    [Arguments]  ${url}  ${username}  ${email}  ${realname}  ${newPassword}  ${comment}
    Go To    ${url}
    Retry Wait Element  ${harbor_span_title}
    Retry Element Click  xpath=${sign_up_for_an_account_xpath}
    Retry Text Input  xpath=${username_xpath}  ${username}
    Retry Text Input  xpath=${email_xpath}  ${email}
    Retry Text Input  xpath=${realname_xpath}  ${realname}
    Retry Text Input  xpath=${newPassword_xpath}  ${newPassword}
    Retry Text Input  xpath=${confirmPassword_xpath}  ${newPassword}
    Retry Text Input  xpath=${comment_xpath}  ${comment}
    Retry Double Keywords When Error  Retry Element Click  ${signup_xpath}  Retry Wait Until Page Not Contains Element  ${signup_xpath}
    Retry Text Input  ${login_name}  ${username}
    Retry Text Input  ${login_pwd}  ${newPassword}
    Retry Double Keywords When Error  Retry Element Click  ${login_btn}  Retry Wait Until Page Not Contains Element  ${login_btn}
    Retry Wait Element  xpath=//span[contains(., '${username}')]
