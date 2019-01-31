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

*** Keywords ***
Change Password
    [Arguments]  ${cur_pw}  ${new_pw}
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/navigator/clr-header/div[3]/clr-dropdown[2]/button/span
    Click Element  xpath=//clr-main-container//clr-dropdown//a[2]
    Sleep  2
    Input Text  xpath=//*[@id='oldPassword']  ${cur_pw}
    Input Text  xpath=//*[@id='newPassword']  ${new_pw}
    Input Text  xpath=//*[@id='reNewPassword']  ${new_pw}
    Sleep  1
    Click Element  xpath=//password-setting/clr-modal//button[2]
    Sleep  2
    Click Element  xpath=${log_xpath}
    Sleep  1

Update User Comment
    [Arguments]  ${new_comment}
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/navigator/clr-header/div[3]/clr-dropdown[2]/button/span
    Click Element  xpath=//clr-main-container//clr-dropdown//a[1]
    Sleep  2
    Input Text  xpath=//*[@id='account_settings_comments']  ${new_comment}
    Sleep  1
    Click Element  xpath=//account-settings-modal/clr-modal//button[2]
    Sleep  2

Logout Harbor
    Wait Until Element Is Visible  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/navigator/clr-header/div[3]/clr-dropdown[2]/button/span
    Wait Until Element Is Enabled  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/navigator/clr-header/div[3]/clr-dropdown[2]/button/span
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/navigator/clr-header/div[3]/clr-dropdown[2]/button/span
    Sleep  2
    Click Link  Log Out
    #Click Element  xpath=//harbor-app/harbor-shell/clr-main-container/navigator/clr-header//clr-dropdown//a[4]
    Sleep  1
    Capture Page Screenshot  Logout.png
    Sleep  2
    Wait Until Keyword Succeeds  5x  1  Page Should Contain Element  xpath=//sign-in//form//*[@class='title']
