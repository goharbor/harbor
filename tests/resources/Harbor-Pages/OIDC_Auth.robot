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

Sign In Harbor With OIDC User
    [Arguments]  ${url}  ${username}=${OIDC_USERNAME}
    ${head_username}=  Set Variable  xpath=//harbor-app/harbor-shell/clr-main-container/navigator/clr-header//clr-dropdown//button[contains(.,'${username}')]
    Init Chrome Driver
    Go To    ${url}
    Retry Element Click    ${log_oidc_provider_btn}
    Retry Text Input    ${dex_login_btn}    ${username}@example.com
    Retry Text Input    ${dex_pwd_btn}    password
    Retry Element Click    ${submit_login_btn}
    Retry Element Click    ${grant_btn}

    #If input box for harbor user name is visible, it means it's the 1st time login of this user,
    #  but if this user has been logged into harbor successfully, this input box will not show up,
    #  so there is condition branch for this stituation.
    ${isVisible}=  Run Keyword And Return Status  Element Should Be Visible  ${oidc_username_input}
    Run Keyword If  '${isVisible}' == 'True'  Run Keywords  Retry Text Input    ${oidc_username_input}    ${username}  AND  Retry Element Click    ${save_btn}
    Retry Wait Element  ${head_username}
