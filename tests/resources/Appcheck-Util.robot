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
Documentation  This resource provides any keywords related to appcheck appliance

*** Variables ***

*** Keywords ***
Sign In Appcheck
    [Arguments]
    ${d}=    Get Current Date    result_format=%m%s
    Go To    ${APPCHECK_URL}
    Retry Text Input  //*[@id='username']  ${APPCHECK_USER}
    Retry Button Click  //button[contains(.,'Continue')]
    Retry Text Input  //*[@id='password']  ${APPCHECK_PWD}
    Retry Button Click  //button[contains(.,'Log in')]
    Retry Wait Element  xpath=//span[contains(., '${APPCHECK_USER}')]
    Retry Wait Element  //div[@id='sidebar']//div[contains(., 'Vulnerability analysis verdict')]
    Sleep  2
    Capture Page Screenshot  /appcheck/appcheck_${d}.png
    Capture Page Screenshot  /appcheck/appcheck.png