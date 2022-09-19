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

*** Keywords ***
Switch To Logs
    Retry Element Click  ${logs_xpath}

Refresh Logs
    Retry Element Click  ${logs_refresh_btn}

Verify Log
    [Arguments]  ${username}  ${resource}  ${resource_type}  ${operation}  ${row_num}=1
    Refresh Logs
    ${real_username}=  Get Text  //clr-datagrid//clr-dg-row[${row_num}]//clr-dg-cell[1]
    ${real_resource}=  Get Text  //clr-datagrid//clr-dg-row[${row_num}]//clr-dg-cell[2]
    ${real_resource_type}=  Get Text  //clr-datagrid//clr-dg-row[${row_num}]//clr-dg-cell[3]
    ${real_operation}=  Get Text  //clr-datagrid//clr-dg-row[${row_num}]//clr-dg-cell[4]
    Should Be Equal  ${real_username}  ${username}
    Should Be Equal  ${real_resource}  ${resource}
    Should Be Equal  ${real_resource_type}  ${resource_type}
    Should Be Equal  ${real_operation}  ${operation}

Verify Log In File
    [Arguments]  ${username}  ${resource}  ${resource_type}  ${operation}  ${audit_log_path}=${log_path}/audit.log
    ${contents}=  OperatingSystem.Get File  ${audit_log_path}
    @{lines}=  Split to lines  ${contents}
    Log  ${lines}[-1]
    @{items}  Create List  operator="${username}"  resource:${resource}  resourceType="${resource_type}"  action:${operation}  time="20
    Should Contain Any  ${lines}[-1]  @{items}