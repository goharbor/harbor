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

Set Pro Create Admin Only	
	#set limit to admin only
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Click Element  xpath=//select[@id="proCreation"]
    Click Element  xpath=//select[@id="proCreation"]//option[@value="adminonly"]
    Click Element  xpath=//config//div/button[1]
	Capture Page Screenshot  AdminCreateOnly.png
	
Set Pro Create Every One	
	#set limit to Every One	
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Click Element  xpath=//select[@id="proCreation"]
    Click Element  xpath=//select[@id="proCreation"]//option[@value="everyone"]
    Click Element  xpath=//config//div/button[1]
    Sleep  2
	Capture Page Screenshot  EveryoneCreate.png

Disable Self Reg	
	Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Mouse Down  xpath=//input[@id="clr-checkbox-selfReg"]
    Mouse Up  xpath=//input[@id="clr-checkbox-selfReg"]
	Sleep  1
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/div/button[1]
	Capture Page Screenshot  DisableSelfReg.png
	Sleep  1

Enable Self Reg	
	Mouse Down  xpath=//input[@id="clr-checkbox-selfReg"]
    Mouse Up  xpath=//input[@id="clr-checkbox-selfReg"]
	Sleep  1
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/div/button[1]
	Capture Page Screenshot  EnableSelfReg.png
	Sleep  1

## System settings	
Switch To System Settings
    Sleep  1
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Click Element  xpath=//config//ul/li[4]
	
Modify Token Expiration
	[Arguments]  ${minutes}
	Input Text  xpath=//input[@id="tokenExpiration"]  ${minutes}
    Click Button  xpath=//config//div/button[1]
	Sleep  1
	
Token Must Be Match
	[Arguments]  ${minutes}
	Textfield Value Should Be  xpath=//input[@id="tokenExpiration"]  ${minutes}