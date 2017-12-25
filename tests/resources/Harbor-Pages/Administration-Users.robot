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
Resource  ../../resources/Util.robot

*** Variables ***
${HARBOR_VERSION}  v1.1.1

*** Keywords ***
Assign User Admin
    [Arguments]  ${user}
    Click Element  xpath=//harbor-user//hbr-filter//clr-icon
    Input Text  xpath=//harbor-user//hbr-filter//input  ${user}
    Sleep  2
    Click Element  xpath=//harbor-user/div/div/h2
    Click Element  xpath=//harbor-user//clr-datagrid//clr-dg-action-overflow
    Click Element  xpath=//harbor-user//clr-dg-action-overflow//button[contains(.,'Admin')]
    Sleep  1

Switch to User Tag
    Click Element  xpath=${administration_user_tag_xpath}
    Sleep  1

Administration Tag Should Display
    Page Should Contain Element  xpath=${administration_tag_xpath}
