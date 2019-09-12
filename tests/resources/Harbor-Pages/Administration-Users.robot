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
Assign User Admin
    [Arguments]  ${user}
    Click Element  xpath=//harbor-user//hbr-filter//clr-icon
    Input Text  xpath=//harbor-user//hbr-filter//input  ${user}
    Sleep  2
    #select checkbox
    Click Element  //clr-dg-row[contains(.,'${user}')]//label
    #click assign admin
    Click Element  //*[@id='set-admin']
    Sleep  1

Switch to User Tag
    Click Element  xpath=${administration_user_tag_xpath}
    Sleep  1

Administration Tag Should Display
    Page Should Contain Element  xpath=${administration_tag_xpath}

User Email Should Exist
    [Arguments]  ${email}
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Switch to User Tag
    Page Should Contain Element  xpath=//clr-dg-cell[contains(., '${email}')]

Add User Button Should Be Disabled
    Sleep  1
    Page Should Contain Element  //button[contains(.,'New') and @disabled='']
