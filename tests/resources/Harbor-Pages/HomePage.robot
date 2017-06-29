# Copyright 2016-2017 VMware, Inc. All Rights Reserved.
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
Library  Selenium2Library
Library  OperatingSystem
Resource  HomePage_Elements.robot

*** Variables ***
${HARBOR_VERSION}  v1.1.1

*** Keywords ***
Sign In Harbor
    [Arguments]  ${url}  ${user}  ${pw}
	Go To    ${url}
    sleep  5
    ${title}=  Get Title
    Log To Console  ${title}
    Should Be Equal  ${title}  Harbor
    Input Text  login_username  ${user}
    Input Text  login_password  ${pw}
    sleep  2
    Click button  css=.btn
    sleep  5
	Log To Console  ${user}
    Wait Until Page Contains  ${user}

Create An New User
    [Arguments]  ${url}  ${username}  ${email}  ${realname}  ${newPassword}  ${comment}
	Go To    ${url}
    sleep  5
    ${title}=  Get Title
    Log To Console  ${title}
    Should Be Equal  ${title}  Harbor
	Capture Page Screenshot
	Sleep  3
    Click Element  xpath=${sign_up_for_an_account_xpath}
    sleep  3
    Input Text  xpath=${username_xpath}  ${username}
    sleep  1
    Input Text  xpath=${email_xpath}  ${email}
    sleep  1
    Input Text  xpath=${realname_xpath}  ${realname}
    sleep  1
    Input Text  xpath=${newPassword_xpath}  ${newPassword}
    sleep  1
    Input Text  xpath=${confirmPassword_xpath}  ${newPassword}
    sleep  1
    Input Text  xpath=${comment_xpath}  ${comment}
    sleep  2
    Click button  xpath=${signup_xpath}
    sleep  5
    Input Text  login_username  ${username}
    Input Text  login_password  ${newPassword}
    sleep  2
    Click button  css=.btn
    sleep  5
    Wait Until Page Contains  ${username}
	sleep  2
