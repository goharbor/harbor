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

*** Variables ***
${HARBOR_VERSION}  v1.1.1

*** Keywords ***
Sign In Harbor
    [Arguments]  ${user}  ${pw}
		Go To    http://10.112.122.5
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
    [Arguments]  ${username}  ${email}  ${realname}  ${newPassword}  ${comment}
		Go To    http://10.112.122.5
    sleep  5
    ${title}=  Get Title
    Log To Console  ${title}
    Should Be Equal  ${title}  Harbor
		Capture Page Screenshot
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/sign-in/div/form/div[1]/a
    sleep  3
    Input Text  xpath=//*[@id="username"]  ${username}
    sleep  1
    Input Text  xpath=//*[@id="email"]  ${email}
    sleep  1
    Input Text  xpath=//*[@id="realname"]  ${realname}
    sleep  1
    Input Text  xpath=//*[@id="newPassword"]  ${newPassword}
    sleep  1
    Input Text  xpath=//*[@id="confirmPassword"]  ${newPassword}
    sleep  1
    Input Text  xpath=//*[@id="comment"]  ${comment}
    sleep  2
    Click button  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/sign-in/sign-up/clr-modal/div/div[1]/div/div[1]/div/div[3]/button[2]
    sleep  5
    Input Text  login_username  ${username}
    Input Text  login_password  ${newPassword}
    sleep  2
    Click button  css=.btn
    sleep  5
    Wait Until Page Contains  ${username}
		sleep  2
