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
Documentation  This resource provides helper functions for docker operations
Resource  Util.robot

*** Keywords ***
Start Selenium Standalone Server Locally
    OperatingSystem.File Should Exist  /go/selenium-server-standalone-3.4.0.jar
    ${handle}=  Start Process  java -jar /go/selenium-server-standalone-3.4.0.jar >./selenium-local.log 2>&1  shell=True
    Process Should Be Running  ${handle}
    Sleep  10s
    [Return]  ${handle}

Init Chrome Driver
    Run  pkill chromedriver
	Run  pkill chrome
	${chrome options}=    Evaluate    sys.modules['selenium.webdriver'].ChromeOptions()    sys
	Call Method    ${chrome options}		add_argument    --headless
	Call Method    ${chrome options}    add_argument    --disable-gpu
    Call Method    ${chrome options}    add_argument    --start-maximized
    Call Method    ${chrome options}    add_argument    --ignore-certificate-errors
	Call Method  	 ${chrome options}    add_argument    --disable-web-security
	Call Method    ${chrome options}    add_argument    --allow-running-insecure-content
	Call Method    ${chrome options}    add_argument    --window-size\=1600,900
	${chrome options.binary_location}    Set Variable    /usr/bin/google-chrome
	Create Webdriver    Chrome    Chrome_headless    chrome_options=${chrome options}
    Sleep  5