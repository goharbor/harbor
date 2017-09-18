// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

*** Settings ***
Resource ../../resources/Uitl.robot
suite setup Start Docker Daemon Locally
default tags regression

*** Test Cases ***
Test Case - Edit authentication
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  admin  Harbor12345
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Click Element  xpath=//select[@id="authMode"]
    Click Element  xpath=//select[@id="authMode"]//option[@value="ldap_auth"]
    Sleep  1
    Input Text  xpath=//input[@id="ldapUrl"]
    Input Text  xpath=//input[@id="ldapSearchDN"]
    Input Text  xpath=//input[@id="ldapSearchPwd"]
    Input Text  xpath=//input[@id="ldapUid"]
    #scope keep subtree
    #click save
    Click Button  xpath=//config//div/button[1]
    Logout Harbor
    #check can change back to db
    Sign In Harbor  admin  Harbor12345
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Page Should Not Contain Element  xpath=//select[@disabled='']
    Logout Harbor
    #signin ldap user
    Sign In Harbor  user001  user001
    Logout Harbor
    #sign in as admin
    Sign In Harbor  admin  Harbor12345
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Page Should Contain Element  xpath=//select[@disabled='']

    #clean database and restart harbor
    Down Harbor
    ${rc} ${output}= Run And Return Rc And Output  rm -rf /data
    Prepare
    Up Harbor

    Create An New User  username=test${d}  email=test${d}@vmware.com  realname=test{d}  newPassword=Test1@34  comment=harbor
    Sign In Harbor  admin  Harbor12345
    Click Element  xpath=//clr-main-containter//nav//ul/li[3]
    Page Should Contain Element  xpath=//select[@disabled='']
    Sleep  1
    Close  Browser
