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
    ...                                    Read Helm Chart
    ...                                    Create Helm Chart Version
    ...                                    Delete Helm Chart Version
    ...                                    Create Helm Chart label
    ...                                    Delete Helm Chart label
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

Create A New System Robot Account
    [Arguments]  ${name}=${null}  ${expiration_type}=default  ${expiration_value}=${null}  ${description}=${null}  ${is_cover_all}=${false}  ${cover_all_permission_list}=@{EMPTY}  ${project_permission_list}=@{EMPTY}
    ${d}=    Get Current Date    result_format=%m%s
    ${name}=  Set Variable If  '${name}'=='${null}'   robot_name${d}  ${name}
    Switch To Robot Account
    Retry Double Keywords When Error  Retry Element Click  ${new_sys_robot_account_btn}  Retry Wait Until Page Contains Element  ${sys_robot_account_name_input}
    Retry Text Input  ${sys_robot_account_name_input}  ${name}
    Run Keyword If  '${expiration_type}' != 'default'  Run Keywords  Retry Element Click  xpath=${sys_robot_account_expiration_type_select}  AND
    ...  Retry Element Click  xpath=${sys_robot_account_expiration_type_select}//option[@value='${expiration_type}']
    Run Keyword If  '${description}' != '${null}'  Retry Text Input  ${sys_robot_account_description_textarea}  ${description}
    Run Keyword If  '${is_cover_all}' == '${true}'  Retry Double Keywords When Error  Retry Element Click  ${sys_robot_account_coverall_chb}   Retry Checkbox Should Be Selected  ${sys_robot_account_coverall_chb_input}
    ...  ELSE  Clear Global Permissions By JavaScript

    # Select project
    FOR  ${project}  IN  @{project_permission_list}
        Log To Console  project: ${project}
        Should Be True    type($project) is not dict
        ${tmp} =    Convert To Dictionary    ${project}
        Should Be True    type($tmp) is dict
        ${project_name}=  Get From Dictionary  ${tmp}  project_name
        Log To Console  project_name: ${project_name}
        ${permission_item_list}=  Get From Dictionary  ${tmp}  permission_item_list
        Log To Console  permission_item_list: ${permission_item_list}
        Filter Project In Project Permisstion List  ${project_name}
        Retry Element Click  //clr-dg-row[contains(.,'${project_name}')]//div[contains(@class,'clr-checkbox-wrapper')]/label
        Retry Element Click  //clr-dg-row[contains(., '${project_name}')]//clr-dropdown/button
        Select Project Permission  ${project_name}  ${permission_item_list}
    END
    # Save it
    Retry Double Keywords When Error  Retry Element Click  ${save_sys_robot_account_btn}  Retry Wait Until Page Not Contains Element  ${save_sys_robot_account_btn}
    Retry Double Keywords When Error  Retry Element Click  ${save_sys_robot_export_to_file_btn}  Retry Wait Until Page Not Contains Element  ${save_sys_robot_export_to_file_btn}
    # Get Robot Account Info
    ${id}  ${name}  ${secret}  ${creation_time}  ${expires_at}=  Get Robot Account Info By File  ${download_directory}/robot$${name}.json
    [Return]  ${name}  ${secret}

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