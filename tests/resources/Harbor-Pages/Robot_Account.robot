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
Create A Random Permission Item List
    ${permission_item_all_list}=  Create List  Push Repository
    ...                                    Pull Repository
    ...                                    Delete Repository
    ...                                    Delete Artifact
    ...                                    Create Tag
    ...                                    Delete Tag
    ...                                    List Tag
    ...                                    Create Artifact label
    ...                                    Delete Artifact label
    ...                                    Create Scan
    ...                                    Stop Scan
    ...                                    List Artifact
    ...                                    List Repository


    Set Suite Variable  ${permission_item_all_list}

    ${len}=  Get Length  ${permission_item_all_list}
    ${tmp_list}=  Create List  @{EMPTY}
    FOR  ${i}  IN RANGE  0  ${${len}-1}
        ${r}=  Evaluate  random.randint(0, 1)
        Run Keyword If  '${r}'=='1'  Append To List  ${tmp_list}  ${permission_item_all_list}[${i}]
    END
    Run Keyword If  ${tmp_list}==@{EMPTY}  Append To List  ${tmp_list}  ${permission_item_all_list}[${0}]
    ${push_pos}=  Get Index From List  ${tmp_list}  ${permission_item_all_list}[${0}]
    ${pull_pos}=  Get Index From List  ${tmp_list}  ${permission_item_all_list}[${1}]
    Run Keyword If  '${push_pos}' >= '${0}' and '${pull_pos}'=='${-1}'  Append To List  ${tmp_list}  ${permission_item_all_list}[${1}]
    [Return]  ${tmp_list}

Create A Random Project Permission List
    [Arguments]  ${project_count}
    ${tmp_list}=  Create List  @{EMPTY}
    FOR  ${i}  IN RANGE  ${project_count}
        ${d}=    Get Current Date    result_format=%m%s
        ${pro_name}=  Set Variable  project_${i}_${d}
        ${permission_item_list}=  Create A Random Permission Item List
        Log To Console  '@{permission_item_list}'
        Create An New Project And Go Into Project  ${pro_name}
        ${tmp_dict} =    Create Dictionary  project_name=${pro_name}  permission_item_list=@{permission_item_list}
        Append To List  ${tmp_list}  ${tmp_dict}
    END
    Log To Console  tmp_list:'@{tmp_list}'
    [Return]  ${tmp_list}

Filter Project In Project Permisstion List
    [Arguments]  ${name}
    Retry Double Keywords When Error  Retry Element Click  ${save_sys_robot_project_filter_chb}  Retry Wait Until Page Contains Element  ${save_sys_robot_project_filter_input}
    Retry Text Input  ${save_sys_robot_project_filter_input}   ${name}
    Retry Double Keywords When Error  Retry Element Click  ${save_sys_robot_project_filter_close_btn}  Retry Wait Until Page Not Contains Element  ${save_sys_robot_project_filter_input}

Clear Global Permissions By JavaScript
    Retry Element Click  //button[contains(., 'RESET PERMISSIONS')]
    FOR  ${i}  IN RANGE  0  10
        Execute JavaScript  document.getElementsByClassName('dropdown-item')[${i}].click();
    END

Select Project Permission
    [Arguments]  ${project_name}  ${permission_item_list}
    FOR  ${permission}  IN  @{permission_item_list}
        Log To Console  project: ${project_name}; permission: ${permission}
        ${item}=  Set Variable  //clr-dg-row[contains(., '${project_name}')]//clr-dropdown/clr-dropdown-menu//span[contains(., '${permission}')]
        Execute JavaScript  document.evaluate("${item}",document.body,null,9,null).singleNodeValue.click();
        Capture Page Screenshot
    END

Create A System Robot Account
    [Arguments]  ${robot_account_name}  ${expiration_type}  ${description}=${null}  ${days}=${null}  ${cover_all_system_resources}=${null}  ${cover_all_project_resources}=${null}
    Retry Element Click  ${new_sys_robot_account_btn}
    Retry Wait Element Should Be Disabled  //button[text()='Next']
    Retry Text Input  ${sys_robot_account_name_input}    ${robot_account_name}
    Run Keyword If  '${description}' != '${null}'  Retry Text Input  ${sys_robot_account_description_textarea}  ${description}
    Select From List By Value  ${sys_robot_account_expiration_type_select}  ${expiration_type}
    Run Keyword If  '${expiration_type}' == 'days'  Retry Text Input  ${sys_robot_account_expiration_input}  ${days}
    Retry Button Click  //button[text()='Next']
    Retry Wait Element Should Be Disabled  ${project_robot_account_create_finish_btn}
    Run Keyword If  '${cover_all_system_resources}' == '${true}'  Retry Element Click  //*[@id='clr-wizard-page-1']//span[text()='Select all']
    Retry Double Keywords When Error  Retry Button Click  //button[text()='Next']  Retry Wait Element Not Visible  //button[text()='Next']
    Run Keyword If  '${cover_all_project_resources}' == '${true}'  Run Keywords  Retry Element Click  ${sys_robot_account_coverall_chb}  AND  Retry Element Click  //*[@id='clr-wizard-page-2']//span[text()='Select all']
    Retry Double Keywords When Error  Retry Element Click  ${project_robot_account_create_finish_btn}  Retry Wait Element Not Visible  ${project_robot_account_create_finish_btn}
    ${robot_account_name}=  Get Text  ${project_robot_account_name_xpath}
    ${token}=  Get Value  //hbr-copy-input//input
    Retry Element Click  //hbr-copy-input//clr-icon
    [Return]  ${robot_account_name}  ${token}

Check System Robot Account API Permission
    [Arguments]  ${robot_account_name}  ${token}  ${admin_user_name}  ${admin_password}  ${resources}  ${expected_status}=0
    ${rc}  ${output}=  Run And Return Rc And Output  USER_NAME='${robot_account_name}' PASSWORD='${token}' ADMIN_USER_NAME=${admin_user_name} ADMIN_PASSWORD=${admin_password} HARBOR_BASE_URL=https://${ip}/api/v2.0 RESOURCES=${resources} python ./tests/apitests/python/test_system_permission.py
    Should Be Equal As Integers  ${rc}  ${expected_status}

System Robot Account Exist
    [Arguments]  ${name}  ${project_count}
    Retry Double Keywords When Error  Retry Element Click  ${filter_dist_btn}  Wait Until Element Is Visible And Enabled  ${filter_dist_input}
    Retry Text Input  ${filter_dist_input}  ${name}
    ${projects}=  Set Variable If  '${project_count}' == 'all'  All projects with  ${project_count} PROJECT
    Retry Wait Until Page Contains Element  //clr-dg-row[contains(.,'${name}') and contains(.,'${projects}')]

Get Robot Account Info By File
    [Arguments]  ${file_path}
    Retry File Should Exist  ${file_path}
    ${json}=  Load Json From File  ${file_path}
    ${id}=  Set Variable  ${json["id"]}
    ${name}=  Set Variable  ${json["name"]}
    ${secret}=  Set Variable  ${json["secret"]}
    ${creation_time}=  Set Variable  ${json["creation_time"]}
    ${expires_at}=  Set Variable  ${json["expires_at"]}
    [Return]  ${id}  ${name}  ${secret}  ${creation_time}  ${expires_at}
