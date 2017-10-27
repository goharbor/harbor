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
    Sleep  2

Set Pro Create Admin Only	
    #set limit to admin only
    Sleep  2
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Sleep  1
    Click Element  xpath=//select[@id="proCreation"]
    Click Element  xpath=//select[@id="proCreation"]//option[@value="adminonly"]
    Sleep  1
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/div/button[1]
    Capture Page Screenshot  AdminCreateOnly.png

Set Pro Create Every One	
    #set limit to Every One	
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Sleep  1
    Click Element  xpath=//select[@id="proCreation"]
    Click Element  xpath=//select[@id="proCreation"]//option[@value="everyone"]
    Sleep  1	
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/div/button[1]
    Sleep  2
    Capture Page Screenshot  EveryoneCreate.png

Disable Self Reg	
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Mouse Down  xpath=${self_reg_xpath}
    Mouse Up  xpath=${self_reg_xpath}
    Sleep  1
    Self Reg Should Be Disabled
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/div/button[1]
    Capture Page Screenshot  DisableSelfReg.png
    Sleep  1

Enable Self Reg	
    Mouse Down  xpath=${self_reg_xpath}
    Mouse Up  xpath=${self_reg_xpath}
    Sleep  1
    Self Reg Should Be Enabled
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/div/button[1]
    Capture Page Screenshot  EnableSelfReg.png
    Sleep  1

Self Reg Should Be Disabled
    Checkbox Should Not Be Selected  xpath=${self_reg_xpath}

Self Reg Should Be Enabled
    Checkbox Should Be Selected  xpath=${self_reg_xpath}

Project Creation Should Display
    Page Should Contain Element  xpath=${project_create_xpath}

Project Creation Should Not Display
    Page Should Not Contain Element  xpath=${project_create_xpath}

## System settings	
Switch To System Settings
    Sleep  1
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Click Element  xpath=//*[@id="config-system"]

Modify Token Expiration
    [Arguments]  ${minutes}
    Input Text  xpath=//*[@id="tokenExpiration"]  ${minutes}
    Click Button  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/div/button[1] 
    Sleep  1

Token Must Be Match
    [Arguments]  ${minutes}
    Textfield Value Should Be  xpath=//*[@id="tokenExpiration"]  ${minutes}

## Replication	
Check Verify Remote Cert	
    Mouse Down  xpath=//*[@id="clr-checkbox-verifyRemoteCert"] 
    Mouse Up  xpath=//*[@id="clr-checkbox-verifyRemoteCert"]
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/div/button[1]
    Capture Page Screenshot  RemoteCert.png
    Sleep  1

Switch To System Replication
    Sleep  1
    Switch To Configure
    Click Element  xpath=//*[@id="config-replication"]
    Sleep  1

Should Verify Remote Cert Be Enabled
    Checkbox Should Not Be Selected  xpath=//*[@id="clr-checkbox-verifyRemoteCert"]

## Email	
Switch To Email
    Switch To Configure
    Click Element  xpath=//*[@id="config-email"]
    Sleep  1

Config Email
    Input Text  xpath=//*[@id="mailServer"]  smtp.vmware.com
    Input Text  xpath=//*[@id="emailPort"]  25
    Input Text  xpath=//*[@id="emailUsername"]  example@vmware.com 
    Input Text  xpath=//*[@id="emailPassword"]  example
    Input Text  xpath=//*[@id="emailFrom"]  example<example@vmware.com>
    Sleep  1    
    Mouse Down  xpath=//*[@id="clr-checkbox-emailSSL"]
    Mouse Up  xpath=//*[@id="clr-checkbox-emailSSL"]
    Sleep  1
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/div/button[1]
    Sleep  6

Verify Email
    Textfield Value Should Be  xpath=//*[@id="mailServer"]  smtp.vmware.com
    Textfield Value Should Be  xpath=//*[@id="emailPort"]  25
    Textfield Value Should Be  xpath=//*[@id="emailUsername"]  example@vmware.com
    Textfield Value Should Be  xpath=//*[@id="emailFrom"]  example<example@vmware.com>
    Checkbox Should Be Selected  xpath=//*[@id="clr-checkbox-emailSSL"]	

Set Scan All To None
    click element  //vulnerability-config//select
    click element  //vulnerability-config//select/option[@value='none']
    sleep  1
    click element  //config//div/button[contains(.,'SAVE')]
Set Scan All To Daily
    click element  //vulnerability-config//select
    click element  //vulnerability-config//select/option[@value='daily']
    sleep  1
    click element  //config//div/button[contains(.,'SAVE')]
Click Scan Now
    click element  //vulnerability-config//button[contains(.,'SCAN')]