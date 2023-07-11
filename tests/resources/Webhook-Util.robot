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
Library  OperatingSystem
Library  Process

*** Keywords ***
Delete All Requests
    Sleep  3
    Run Keyword And Ignore Error  Click button  //button[contains(., 'Delete all requests')]

Verify Request
    [Arguments]  &{property}
    FOR  ${key}  IN  @{property.keys()}
        Wait Until Page Contains  "${key}":"${property['${key}']}"
    END

Get Latest Webhook Execution ID
    ${execution_id}=  Get Text  //clr-dg-row[1]//clr-dg-cell[1]//a
    [Return]  ${execution_id}

Verify Webhook Execution
    [Arguments]  ${execution_id}  ${vendor_type}  ${status}  ${event_type}  ${payload_data}
    Retry Wait Until Page Contains Element  //clr-dg-row[.//clr-dg-cell/a[text()=${execution_id}]]//clr-dg-cell[3][contains(.,'${status}')]
    Wait Until Page Contains Element  //clr-dg-row[.//clr-dg-cell/a[text()=${execution_id}]]//clr-dg-cell[2][contains(.,'WEBHOOK')]
    Wait Until Page Contains Element  //clr-dg-row[.//clr-dg-cell/a[text()=${execution_id}]]//clr-dg-cell[4][contains(.,'${event_type}')]
    Retry Element Click  //clr-dg-row[.//clr-dg-cell/a[text()=${execution_id}]]//clr-dg-cell[5]
    FOR  ${key}  IN  @{payload_data.keys()}
        Wait Until Page Contains  "${key}": "${payload_data['${key}']}"
    END

Verify Webhook Execution Log
    [Arguments]  ${execution_id}  ${log}=success to run webhook job
    Retry Link Click  //clr-dg-row//clr-dg-cell/a[text()=${execution_id}]
    Retry Link Click  //clr-dg-row[1]//clr-dg-cell[5]//a
    Switch Window  locator=NEW
    Wait Until Page Contains  ${log}
