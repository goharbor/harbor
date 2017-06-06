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
Init LDAP
    ${rc}  ${output}=  Run And Return Rc And Output  ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'
    Log  ${output}
    Sleep  2
    Input Text  xpath=//*[@id="ldapUrl"]  ldap://${output}
    Sleep  1
    Input Text  xpath=//*[@id="ldapSearchDN"]  cn=admin,dc=example,dc=org
    Sleep  1
    Input Text  xpath=//*[@id="ldapSearchPwd"]  admin
    Sleep  1
    Input Text  xpath=//*[@id="ldapBaseDN"]  dc=example,dc=org
    Sleep  1
    Input Text  xpath=//*[@id="ldapUid"]  cn
    Sleep  1
    Capture Page Screenshot
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/div/button[1]
    Sleep  2
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/div/button[3]
    Sleep  1
    Capture Page Screenshot

Switch To Configure
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/nav/section/section/ul/li[3]/a
    Sleep  1
