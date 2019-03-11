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
    ${capabilities}=    Evaluate    sys.modules['selenium.webdriver'].DesiredCapabilities.CHROME    sys
    Set To Dictionary    ${capabilities}    acceptInsecureCerts    ${True}
    Call Method    ${chrome options}    add_argument    --headless
    Call Method    ${chrome options}    add_argument    --disable-gpu
    Call Method    ${chrome options}    add_argument    --start-maximized
    Call Method    ${chrome options}    add_argument    --no-sandbox
    Call Method    ${chrome options}    add_argument    --window-size\=1600,900
    ${chrome options.binary_location}    Set Variable    /usr/bin/google-chrome
    #Create Webdriver    Chrome    Chrome_headless    chrome_options=${chrome options}    desired_capabilities=${capabilities}
    :For  ${n}  IN RANGE  1  6
    \    Log To Console  Trying Create Webdriver ${n} times ...
    \    ${out}  Run Keyword And Ignore Error  Create Webdriver    Chrome    Chrome_headless    chrome_options=${chrome options}    desired_capabilities=${capabilities}
    \    Log To Console  Return value is ${out[0]}
    \    Exit For Loop If  '${out[0]}'=='PASS'
    \    Sleep  2
    Run Keyword If  '${out[0]}'=='FAIL'  Capture Page Screenshot
    Should Be Equal As Strings  '${out[0]}'  'PASS'
    Sleep  5
