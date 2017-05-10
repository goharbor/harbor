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

*** Variables ***
${HARBOR_VERSION}  v1.1.1

*** Keywords ***
Install Harbor to Test Server
		Log To Console  \nStart Docker Daemon
		Start Docker Daemon Locally
    Log To Console  \nconfig harbor cfg
    Run Keywords  Config Harbor cfg
		Run Keywords  Prepare Cert
    Log To Console  \ncomplile and up harbor now
    Run Keywords  Compile and Up Harbor With Source Code
    ${rc}  ${output}=  Run And Return Rc And Output  docker ps
    Should Be Equal As Integers  ${rc}  0
    Log To Console  \n${output}

Config Harbor cfg
    # Will change the IP and Protocol in the harbor.cfg
    [Arguments]  ${http_proxy}=https
    ${rc}  ${output}=  Run And Return Rc And Output  ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'
    Log  ${output}
    ${rc}=  Run And Return Rc  sed "s/reg.mydomain.com/${output}/" -i ./make/harbor.cfg
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  sed "s/^ui_url_protocol = .*/ui_url_protocol = ${http_proxy}/g" -i ./make/harbor.cfg
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0

Prepare Cert
    # Will change the IP and Protocol in the harbor.cfg
		${rc}=  Run And Return Rc  ./tests/generateCerts.sh
		Log  ${rc}
		Should Be Equal As Integers  ${rc}  0

Compile and Up Harbor With Source Code
    [Arguments]  ${golang_image}=golang:1.7.3  ${clarity_image}=vmware/harbor-clarity-ui-builder:0.8.4  ${with_notary}=true
    ${rc}  ${output}=  Run And Return Rc And Output  make install GOBUILDIMAGE=${golang_image} COMPILETAG=compile_golangimage CLARITYIMAGE=${clarity_image} NOTARYFLAG=${with_notary} HTTPPROXY=
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    Sleep  30

Sign In Harbor
    [Arguments]  ${user}  ${pw}
    ${chrome_switches} =         Create List          enable-logging       v=1
    ${desired_capabilities} =    Create Dictionary    chrome.switches=${chrome_switches}     platform=LINUX     phantomjs.binary.path=/go/phantomjs
    Open Browser  url=http://localhost  browser=PhantomJS  remote_url=http://127.0.0.1:4444/wd/hub  desired_capabilities=${desired_capabilities}
    Set Window Size  1280  1024
    sleep  10
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
    ${chrome_switches} =         Create List          enable-logging       v=1
    ${desired_capabilities} =    Create Dictionary    chrome.switches=${chrome_switches}     platform=LINUX     phantomjs.binary.path=/go/phantomjs
    Open Browser  url=http://localhost  browser=PhantomJS  remote_url=http://127.0.0.1:4444/wd/hub  desired_capabilities=${desired_capabilities}
    Set Window Size  1920  1080
    sleep  10
    ${title}=  Get Title
    Log To Console  ${title}
    Should Be Equal  ${title}  Harbor
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
    Click button  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/sign-in/sign-up/clr-modal/div/div[1]/div/div[3]/button[2]
    sleep  5
    Input Text  login_username  ${username}
    Input Text  login_password  ${newPassword}
    sleep  2
    Click button  css=.btn
    sleep  5
    Wait Until Page Contains  ${username}
		Close Browser

Logout Harbor
		Wait Until Element Is Visible  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/navigator/clr-header/div[3]/clr-dropdown[2]/button/span
		Wait Until Element Is Enabled  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/navigator/clr-header/div[3]/clr-dropdown[2]/button/span
    Click Button  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/navigator/clr-header/div[3]/clr-dropdown[2]/button
    sleep  2
		Capture Page Screenshot
		sleep  2
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/navigator/clr-header/div[3]/clr-dropdown[2]/div/a[4]
    Wait Until Keyword Succeeds  5x  1  Page Should Contain Element  xpath=//*[@id="pop_repo"]/top-repo/div/div[1]/h3

Create An New Project
		[Arguments]  ${projectname}
		sleep  1
		Click Button  css=.btn
		sleep  1
		Log To Console  Project Name: ${projectname}
		Input Text  css=#create_project_name  ${projectname}
		Click Element  css=html body.no-scrolling harbor-app harbor-shell clr-main-container.main-container div.content-container div.content-area.content-area-override project div.row div.col-lg-12.col-md-12.col-sm-12.col-xs-12 div.row.flex-items-xs-between div.option-left create-project clr-modal div.modal div.modal-dialog div.modal-content div.modal-footer button.btn.btn-primary
		sleep  2
		Capture Page Screenshot
		Wait Until Page Contains  ${projectname}
		Wait Until Page Contains  Project Admin
		Capture Page Screenshot
