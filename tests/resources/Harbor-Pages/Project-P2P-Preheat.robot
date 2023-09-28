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
Switch To P2P Preheat
    Retry Element Click  xpath=${project_p2p_preheat__tag_xpath}

Select Distribution For P2P Preheat
    [Arguments]  ${provider}
    Retry Element Click  ${p2p_preheat_provider_select_id}
    Retry Element Click  ${p2p_preheat_provider_select_id}//option[contains(.,'${provider}')]

Select P2P Preheat Policy
    [Arguments]  ${name}
    Retry Element Click  //clr-dg-row[contains(.,'${name}')]//clr-radio-wrapper/label[contains(@class,'clr-control-label')]

P2P Preheat Policy Exist
    [Arguments]  ${name}  ${repo}=${null}
    ${policy_row_xpath}=  Set Variable If  '${repo}'=='${null}'  //clr-dg-row[contains(.,'${name}')]  //clr-dg-row[contains(.,'${name}') and contains(.,'${repo}')]
    Retry Wait Until Page Contains Element  ${policy_row_xpath}

P2P Preheat Policy Not Exist
    [Arguments]  ${name}
    Retry Wait Until Page Not Contains Element  //clr-dg-row[contains(.,'${name}')]

Create An New P2P Preheat Policy
    [Arguments]  ${policy_name}  ${dist_name}  ${repo}  ${tag}  ${trigger_type}=Manual  ${schedule_type}=${null}  ${schedule_cron}=${null}
    Switch To P2P Preheat
    Retry Element Click  ${p2p_preheat_new_policy_btn_id}
    Select Distribution For P2P Preheat  ${dist_name}
    Retry Text Input  ${p2p_preheat_name_input_id}  ${policy_name}
    Retry Text Input  ${p2p_preheat_repoinput_id}  ${repo}
    Retry Text Input  ${p2p_preheat_tag_input_id}  ${tag}
    Select P2P Preheat Policy Trigger  ${trigger_type}
    Run Keyword If  '${trigger_type}' == 'Scheduled'  Retry Element Click  ${p2p_preheat_scheduled_edit_id}
    Run Keyword If  '${trigger_type}' == 'Scheduled'  Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_scheduled_type_select_id}  Retry Element Click  ${p2p_preheat_scheduled_type_select_id}//option[contains(.,'${schedule_type}')]
    Run Keyword If  '${schedule_type}' == 'Custom'  Retry Text Input  ${p2p_preheat_scheduled_cron_input_id}  ${schedule_cron}
    Run Keyword If  '${trigger_type}' == 'Scheduled'  Retry Element Click  ${p2p_preheat_scheduled_save_btn_xpath}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_add_save_btn_id}  Retry Wait Until Page Not Contains Element  xpath=${p2p_preheat_add_save_btn_id}
    P2P Preheat Policy Exist  ${policy_name}

Select P2P Preheat Policy Trigger
    [Arguments]  ${mode}
    Retry Element Click  ${p2p_preheat_trigger_select}
    Retry Element Click  ${p2p_preheat_trigger_select}//option[contains(.,'${mode}')]

Edit A P2P Preheat Policy
    [Arguments]  ${name}  ${repo}  ${trigger_type}=${null}
    Switch To P2P Preheat
    Retry Double Keywords When Error  Select P2P Preheat Policy   ${name}  Wait Until Element Is Visible  ${p2p_execution_header}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_action_btn_id}  Wait Until Element Is Visible And Enabled  ${p2p_preheat_edit_btn_id}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_edit_btn_id}  Wait Until Element Is Visible And Enabled  ${p2p_preheat_name_input_id}
    Retry Text Input  ${p2p_preheat_repoinput_id}  ${repo}
    Run Keyword If  '${trigger_type}' != '${null}'  Select P2P Preheat Policy Trigger  ${trigger_type}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_edit_save_btn_id}  Retry Wait Until Page Not Contains Element  xpath=${p2p_preheat_edit_save_btn_id}
    P2P Preheat Policy Exist  ${name}  repo=${repo}

Set P2P Preheat Policy Schedule
    [Arguments]  ${name}  ${type}  ${cron}=${null}
    Retry Double Keywords When Error  Select P2P Preheat Policy   ${name}  Wait Until Element Is Visible  ${p2p_execution_header}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_action_btn_id}  Wait Until Element Is Visible And Enabled  ${p2p_preheat_edit_btn_id}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_edit_btn_id}  Wait Until Element Is Visible And Enabled  ${p2p_preheat_name_input_id}
    Retry Double Keywords When Error  Select P2P Preheat Policy Trigger  Scheduled  Retry Element Click  ${p2p_preheat_scheduled_edit_id}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_scheduled_type_select_id}  Retry Element Click  ${p2p_preheat_scheduled_type_select_id}//option[contains(.,'${type}')]
    Run Keyword If  '${type}' == 'Custom'  Retry Text Input  ${p2p_preheat_scheduled_cron_input_id}  ${cron}
    Retry Element Click  ${p2p_preheat_scheduled_save_btn_xpath}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_edit_save_btn_id}  Retry Wait Until Page Not Contains Element  xpath=${p2p_preheat_edit_save_btn_id}
    P2P Preheat Policy Exist  ${name}

Delete A P2P Preheat Policy
    [Arguments]  ${name}
    Switch To P2P Preheat
    Retry Double Keywords When Error  Select P2P Preheat Policy  ${name}  Wait Until Element Is Visible  ${p2p_execution_header}
    Retry Wait Until Page Not Contains Element  //clr-datagrid[contains(.,'ID')]//div//clr-dg-row[1]//clr-dg-cell[2][text()=' Running ']
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_action_btn_id}  Wait Until Element Is Visible And Enabled  ${p2p_preheat_del_btn_id}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_del_btn_id}  Wait Until Element Is Visible And Enabled  ${delete_confirm_btn}
    Retry Double Keywords When Error  Retry Element Click  ${delete_confirm_btn}  Retry Wait Until Page Not Contains Element  ${delete_confirm_btn}
    P2P Preheat Policy Not Exist  ${name}

Execute P2P Preheat And Verify
    [Arguments]  ${project_name}  ${policy_name}  ${contain}  ${not_contain}
    Go Into Project  ${project_name}
    Switch To P2P Preheat
    Execute P2P Preheat  ${policy_name}
    Verify Latest Execution Result  ${project_name}  ${policy_name}  ${contain}  ${not_contain}

Execute P2P Preheat
    [Arguments]  ${name}  ${expected_status}=Success
    Retry Double Keywords When Error  Select P2P Preheat Policy  ${name}  Wait Until Element Is Visible  ${p2p_execution_header}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_action_btn_id}  Wait Until Element Is Visible And Enabled  ${p2p_preheat_execute_btn_id}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_execute_btn_id}  Wait Until Element Is Visible And Enabled  ${p2p_preheat_confirm_execute_btn_id}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_confirm_execute_btn_id}  Wait Until Element Is Visible And Enabled  //clr-datagrid//div//clr-dg-row[1]//clr-dg-cell[2][contains(.,'${expected_status}')]

Verify Latest Execution Result
    [Arguments]  ${project_name}  ${policy_name}  ${contain}  ${not_contain}=${null}  ${expected_status}=Success
    Retry Double Keywords When Error  Select P2P Preheat Policy  ${policy_name}  Wait Until Element Is Visible  ${p2p_preheat_executions_refresh_xpath}
    Retry Keyword N Times When Error  5  Retry P2P Preheat Be Successful  ${project_name}  ${policy_name}  ${contain}  ${not_contain}

Retry P2P Preheat Be Successful
    [Arguments]  ${project_name}  ${policy_name}  ${contain}  ${not_contain}=${null}  ${expected_status}=Success
    Retry Element Click  ${p2p_preheat_executions_refresh_xpath}
    ${latest_execution_id}=  Get Text  ${p2p_preheat_latest_execute_id_xpath}
    P2P Preheat Be Successful  ${project_name}  ${policy_name}  ${latest_execution_id}  ${contain}  ${not_contain}  ${expected_status}

P2P Preheat Be Successful
    [Arguments]  ${project_name}  ${policy_name}  ${execution_id}  ${contain}  ${not_contain}=${null}  ${expected_status}=Success
    ${rc}  ${output}=  Run And Return Rc And Output  curl -u ${HARBOR_ADMIN}:${HARBOR_PASSWORD} -i --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/v2.0/projects/${project_name}/preheat/policies/${policy_name}/executions/${execution_id}/tasks"
    Log All  ${output}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  ${expected_status}
    Should Contain  ${output}  @{contain}
    ${out}  Run Keyword And Ignore Error  Get Length  ${not_contain}
    Run Keyword If  '${out[0]}'=='PASS'  Should Not Contain Any  ${output}  @{not_contain}

Verify Artifact Is Pushed Event
    [Arguments]  ${project_name}  ${policy_name}  ${image}  ${tag}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image}  ${tag}  ${tag}
    Go Into Project  ${project_name}
    Switch To P2P Preheat
    ${contain}  Create List  ${project_name}/${image}:${tag}
    Verify Latest Execution Result  ${project_name}  ${policy_name}  ${contain}

Verify Artifact Is Scanned Event
    [Arguments]  ${project_name}  ${policy_name}  ${image}  ${tag}
    Go Into Repo  ${project_name}  ${image}
    Scan Repo  ${tag}  Succeed
    Back Project Home  ${project_name}
    Switch To P2P Preheat
    ${contain}  Create List  ${project_name}/${image}:${tag}
    Verify Latest Execution Result  ${project_name}  ${policy_name}  ${contain}

Verify Artifact Is Labeled Event
    [Arguments]  ${project_name}  ${policy_name}  ${image}  ${tag}  ${label}
    Go Into Project  ${project_name}
    Switch To Project Label
    Create New Labels  ${label}
    Go Into Repo  ${project_name}  ${image}
    Add Labels To Tag  ${tag}  ${label}
    Back Project Home  ${project_name}
    Switch To P2P Preheat
    ${contain}  Create List  ${project_name}/${image}:${tag}
    Verify Latest Execution Result  ${project_name}  ${policy_name}  ${contain}

Get P2P Preheat Logs
    [Arguments]  ${project_name}  ${policy_name}
    ${cmd}=  Set Variable  curl -u ${HARBOR_ADMIN}:${HARBOR_PASSWORD} -s --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/v2.0/projects/${project_name}/preheat/policies/${policy_name}/executions"
    Log All  cmd:${cmd}
    ${rc}  ${output}=  Run And Return Rc And Output  ${cmd}
    Log All  ${output}
    [Return]  ${output}
