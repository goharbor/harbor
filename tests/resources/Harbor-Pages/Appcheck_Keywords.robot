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
Sign In Appcheck
    [Arguments]  ${url}  ${user}  ${pw}
    Go To  https://appcheck.eng.vmware.com/products/97277/
    #Go To    ${url}
    #Retry Text Input //*[@id='username']  ${user}
    #Retry Text Input  //*[@id='password']  ${pw}
    Retry Text Input  //*[@id='username']  danfengl
    Retry Button Click  //button[contains(.,'Continue')]
    Retry Text Input  //*[@id='password']  Teligen9898!@#$%
    Retry Button Click  //button[contains(.,'Log in')]
    #Retry Wait Element  xpath=//span[contains(., '${user}')]
    Retry Wait Element  xpath=//span[contains(., 'danfengl')]
    Retry Wait Element  //div[@id='sidebar']//div[contains(., 'Vulnerability analysis verdict')]
    Sleep  2
    Capture Page Screenshot