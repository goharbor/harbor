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
${HARBOR_VERSION}  v1.1.1
${timeout}  60
${login_btn}  css=.btn
${harbor_span_title}  xpath=//span[contains(., 'Harbor')]
${login_name}  id=login_username
${login_pwd}  id=login_password
*** Keywords ***
Sign In Harbor
    [Arguments]  ${url}  ${user}  ${pw}
    Go To    ${url}
    Wait Until Element Is Enabled  ${harbor_span_title}
    Wait Until Element Is Visible  ${login_name}
    Wait Until Element Is Visible  ${login_pwd}
    Input Text  ${login_name}  ${user}
    Input Text  ${login_pwd}  ${pw}
    Wait Until Element Is Visible  ${login_btn}
    Click button  ${login_btn}
    Log To Console  ${user}
    Wait Until Element Is Visible  xpath://span[contains(., '${user}')]

Capture Screenshot And Source
    Capture Page Screenshot
    Log Source

Sign Up Should Not Display
    Page Should Not Contain Element  xpath=${sign_up_button_xpath}

Create An New User
    [Arguments]  ${url}  ${username}  ${email}  ${realname}  ${newPassword}  ${comment}
    Go To    ${url}
    Wait Until Element Is Visible  ${harbor_span_title}  timeout=${timeout}
    Wait Until Element Is Visible  xpath=${sign_up_for_an_account_xpath}  timeout=${timeout}
    Click Element  xpath=${sign_up_for_an_account_xpath}
    Wait Until Element Is Visible  xpath=${username_xpath}  timeout=${timeout}
    Input Text  xpath=${username_xpath}  ${username}
    Wait Until Element Is Visible  xpath=${email_xpath}  timeout=${timeout}
    Input Text  xpath=${email_xpath}  ${email}
    Wait Until Element Is Visible  xpath=${realname_xpath}  timeout=${timeout}
    Input Text  xpath=${realname_xpath}  ${realname}
    Wait Until Element Is Visible  xpath=${newPassword_xpath}  timeout=${timeout}
    Input Text  xpath=${newPassword_xpath}  ${newPassword}
    Wait Until Element Is Visible  xpath=${confirmPassword_xpath}  timeout=${timeout}
    Input Text  xpath=${confirmPassword_xpath}  ${newPassword}
    Wait Until Element Is Visible  xpath=${comment_xpath}  timeout=${timeout}
    Input Text  xpath=${comment_xpath}  ${comment}
    Wait Until Element Is Visible  xpath=${signup_xpath}  timeout=${timeout}
    Click button  xpath=${signup_xpath}
    Sleep  2
    Wait Until Element Is Visible  ${login_name}  timeout=${timeout}
    Input Text  ${login_name}  ${username}
    Wait Until Element Is Visible  ${login_pwd}  timeout=${timeout}
    Input Text  ${login_pwd}  ${newPassword}
    Wait Until Element Is Visible  ${login_btn}  timeout=${timeout}
    Click button  ${login_btn}
    Wait Until Element Is Visible  xpath=//span[contains(., '${username}')]  timeout=${timeout}
