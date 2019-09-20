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
    Page Should Not Contain Element  xpath=${sign_up_button_xpath}

Create An New User
    [Arguments]  ${url}  ${username}  ${email}  ${realname}  ${newPassword}  ${comment}
    Go To    ${url}
    Wait Until Element Is Visible  ${harbor_span_title}
    Wait Until Element Is Visible  xpath=${sign_up_for_an_account_xpath}
    Click Element  xpath=${sign_up_for_an_account_xpath}
    Wait Until Element Is Visible  xpath=${username_xpath}
    Input Text  xpath=${username_xpath}  ${username}
    Wait Until Element Is Visible  xpath=${email_xpath}
    Input Text  xpath=${email_xpath}  ${email}
    Wait Until Element Is Visible  xpath=${realname_xpath}
    Input Text  xpath=${realname_xpath}  ${realname}
    Wait Until Element Is Visible  xpath=${newPassword_xpath}
    Input Text  xpath=${newPassword_xpath}  ${newPassword}
    Wait Until Element Is Visible  xpath=${confirmPassword_xpath}
    Input Text  xpath=${confirmPassword_xpath}  ${newPassword}
    Wait Until Element Is Visible  xpath=${comment_xpath}
    Input Text  xpath=${comment_xpath}  ${comment}
    Wait Until Element Is Visible  xpath=${signup_xpath}
    Click button  xpath=${signup_xpath}
    Sleep  2
    Wait Until Element Is Visible  ${login_name}
    Input Text  ${login_name}  ${username}
    Wait Until Element Is Visible  ${login_pwd}
    Input Text  ${login_pwd}  ${newPassword}
    Wait Until Element Is Visible  ${login_btn}
    Click button  ${login_btn}
    Wait Until Element Is Visible  xpath=//span[contains(., '${username}')]
