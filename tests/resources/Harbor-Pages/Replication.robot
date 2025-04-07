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
Filter Replication Rule
    [Arguments]  ${ruleName}  ${exist}=${true}
    ${rule_name_element}=  Set Variable  xpath=//clr-dg-cell[contains(.,'${ruleName}')]
    Retry Element Click  ${filter_rules_btn}
    Retry Clear Element Text  ${filter_rules_input}
    Retry Text Input  ${filter_rules_input}  ${ruleName}
    Run Keyword If  ${exist}==${true}  Retry Wait Until Page Contains Element   ${rule_name_element}
    ...  ELSE  Retry Wait Element  xpath=//clr-dg-placeholder[contains(.,\"We couldn\'t find any replication rules!\")]

Filter Registry
    [Arguments]  ${registry_name}
    ${registry_name_element}=  Set Variable  xpath=//clr-dg-cell[contains(.,'${registry_name}')]
    Retry Element Click  ${filter_registry_btn}
    Retry Text Input  ${filter_registry_input}  ${registry_name}
    Retry Wait Until Page Contains Element   ${registry_name_element}

Select Dest Registry
    [Arguments]    ${endpoint}
    Retry Element Click    ${dest_registry_dropdown_list}
    Retry Element Click    ${dest_registry_dropdown_list}//option[contains(.,'${endpoint}')]

Select Source Registry
    [Arguments]    ${endpoint}
    Retry Element Click    ${src_registry_dropdown_list}
    Retry Element Click    ${src_registry_dropdown_list}//option[contains(.,'${endpoint}')]

Select Filter Tag Model
    [Arguments]    ${type}
    Retry Element Click  ${filter_tag_model_select}
    Retry Element Click  ${filter_tag_model_select}//option[contains(.,'${type}')]

Select Filter Label Model
    [Arguments]    ${type}
    Retry Element Click  ${filter_label_model_select}
    Retry Element Click  ${filter_label_model_select}//option[contains(.,'${type}')]

Select Filter Label
    [Arguments]    @{labels}
    Retry Element Click  ${filter_label_xpath}
    FOR  ${label}  IN  @{labels}
        Log  ${label}
        Retry Element Click  //hbr-label-piece//span[contains(text(), '${label}')]
    END
    Retry Element Click  ${filter_label_xpath}

Select Bandwidth Unit
    [Arguments]    ${unit}
    Retry Element Click  ${bandwidth_unit_select}
    Retry Element Click  ${bandwidth_unit_select}//option[contains(.,'${unit}')]

Select flattening
    [Arguments]    ${type}
    Retry Element Click    ${flattening_select}
    Retry Element Click    ${flattening_select}//option[contains(.,'${type}')]

Select Trigger
    [Arguments]    ${mode}
    Retry Element Click    ${rule_trigger_select}
    Retry Element Click    ${rule_trigger_select}//option[contains(.,'${mode}')]

Select Destination URL
    [Arguments]    ${type}
    Retry Element Click  ${destination_url_xpath}
    Retry Element Click  //div[contains(@class, 'selectBox')]//li[contains(.,'${type}')]

Check New Rule UI Without Endpoint
    Retry Element Click    ${new_replication-rule_button}
    Page Should Contain    Please add an endpoint first
    Retry Element Click    ${link_to_registries}
    Retry Wait Until Page Contains    Endpoint URL
    Retry Wait Element  ${new_endpoint_button}

Create A New Endpoint
    [Arguments]    ${provider}    ${name}    ${url}    ${username}    ${pwd}    ${save}=Y
    #click new button
    Retry Element Click  xpath=${new_endpoint_button}
    #input necessary info
    Select From List By Value  ${provider_selector}  ${provider}
    Retry Text Input  xpath=${destination_name_xpath}    ${name}
    Run Keyword If  '${provider}' == 'harbor' or '${provider}' == 'gitlab'  Run keyword  Retry Text Input  xpath=${destination_url_xpath}  ${url}
    Run Keyword If  '${provider}' == 'aws-ecr' or '${provider}' == 'google-gcr'  Run keyword  Select Destination URL  ${url}
    Run Keyword If  '${provider}' != 'google-gcr' and '${username}' != '${null}'  Retry Password Input  xpath=${destination_username_xpath}  ${username}
    Run Keyword If  '${username}' != '${null}'  Retry Text Input  xpath=${destination_username_xpath}  ${username}
    Run Keyword If  '${pwd}' != '${null}'  Retry Password Input  xpath=${destination_password_xpath}  ${pwd}
    #cancel verify cert since we use a selfsigned cert
    Retry Element Click  ${destination_insecure_xpath}
    Run Keyword If  '${save}' == 'Y'  Run keyword  Retry Double Keywords When Error  Retry Element Click  ${replication_save_xpath}  Retry Wait Until Page Not Contains Element  ${replication_save_xpath}
    Run Keyword If  '${save}' == 'Y'  Run keyword  Filter Registry  ${name}
    Run Keyword If  '${save}' == 'N'  No Operation

Create A Rule With Existing Endpoint
    [Arguments]  ${name}  ${replication_mode}  ${filter_project_name}  ${resource_type}  ${endpoint}  ${dest_namespace}
    ...    ${mode}=Manual  ${cron}="* */59 * * * *"  ${del_remote}=${false}  ${filter_tag}=${false}  ${filter_tag_model}=matching  ${filter_label}=${false}  ${filter_label_model}=matching
    ...    ${flattening}=Flatten 1 Level  ${bandwidth}=-1  ${bandwidth_unit}=Kbps  ${copy_by_chunk}=${false}
    # Click new
    Retry Double Keywords When Error  Retry Element Click  ${new_name_xpath}  Wait Until Element Is Enabled  ${rule_name}
    # Input name
    Retry Text Input  ${rule_name}  ${name}
    Run Keyword If  '${replication_mode}' == 'push'  Run Keywords  Retry Element Click  ${replication_mode_radio_push}  AND  Select Dest Registry  ${endpoint}
    ...    ELSE  Run Keywords  Retry Element Click  ${replication_mode_radio_pull}  AND  Select Source Registry  ${endpoint}
    # Set filter
    Retry Password Input  ${filter_name_id}  ${filter_project_name}
    Run Keyword If  '${filter_tag_model}' != 'matching'  Select Filter Tag Model  ${filter_tag_model}
    Run Keyword If  '${filter_tag}' != '${false}'  Retry Text Input    ${filter_tag_id}    ${filter_tag}
    Run Keyword If  '${filter_label_model}' != 'matching'  Select Filter Label Model  ${filter_label_model}
    Run Keyword If  '${filter_label}' != '${false}'  Select Filter Label  ${filter_label}
    Run Keyword And Ignore Error    Select From List By Value    ${rule_resource_selector}    ${resource_type}
    Retry Text Input    ${dest_namespace_xpath}    ${dest_namespace}
    Select flattening  ${flattening}
    # Set trigger
    Select Trigger  ${mode}
    Run Keyword If  '${mode}' == 'Scheduled'  Retry Text Input  ${targetCron_id}  ${cron}
    Run Keyword If  '${mode}' == 'Event Based' and '${del_remote}' == '${true}'  Retry Element Click  ${del_remote_checkbox}
    # Set bandwidth
    Run Keyword If  '${bandwidth}' != '-1'  Retry Text Input  ${bandwidth_input}  ${bandwidth}
    Run Keyword If  '${bandwidth_unit}' != 'Kbps'  Select Bandwidth Unit  ${bandwidth_unit}
    # Set copy by chunk
    Run Keyword If  '${copy_by_chunk}' == '${true}'  Retry Element Click  ${copy_by_chunk_checkbox}
    # Click save
    Retry Double Keywords When Error  Retry Element Click  ${rule_save_button}  Retry Wait Until Page Not Contains Element  ${rule_save_button}

Endpoint Is Unpingable
    Retry Element Click  ${ping_test_button}
    Wait Until Page Contains  Failed

Endpoint Is Pingable
    Retry Element Click  ${ping_test_button}
    Wait Until Page Contains  successfully

Disable Certificate Verification
    Checkbox Should Be Selected  ${destination_insecure_checkbox}
    Retry Element Click  ${destination_insecure_xpath}

Enable Certificate Verification
    Checkbox Should Not Be Selected  ${destination_insecure_checkbox}
    Retry Element Click  ${destination_insecure_xpath}

Switch To Registries
    Retry Element Click  ${nav_to_registries}

Switch To Replication Manage
    Retry Element Click  ${nav_to_replications}

Trigger Replication Manual
    [Arguments]  ${rule}
    Retry Element Click  ${rule_filter_search}
    Retry Text Input   ${rule_filter_input}  ${rule}
    Retry Element Click  //clr-dg-row[contains(.,'${rule}')]//label[contains(@class,'clr-control-label')]
    Retry Element Click  ${action_bar_replicate}
    Retry Wait Until Page Contains Element  ${dialog_replicate}
    #change from click to mouse down and up
    Mouse Down  ${dialog_replicate}
    Mouse Up  ${dialog_replicate}
    Retry Wait Until Page Contains Element  //*[@id='contentAll']//div[contains(.,'${rule}')]/../div/clr-icon[@shape='success-standard']

Rename Rule
    [Arguments]  ${rule}  ${newname}
    Retry Element Click  ${rule_filter_search}
    Retry Text Input  ${rule_filter_input}  ${rule}
    Retry Element Click  //clr-dg-row[contains(.,'${rule}')]//label[contains(@class,'clr-control-label')]
    Retry Element Click  ${replication_rule_action}
    Retry Element Click  ${replication_rule_action_bar_edit}
    Retry Text Input  ${rule_name}  ${newname}
    Retry Element Click  ${rule_save_button}

Select Rule
    [Arguments]  ${rule}
    Retry Double Keywords When Error  Retry Element Click  //clr-dg-row[contains(.,'${rule}')]/div/div[1]/div  Element Should Be Enabled  ${replication_rule_exec_id}

Stop Jobs
    Retry Element Click  ${stop_jobs_button}

View Job Log
    [arguments]  ${job}
    Retry Element Click  ${job_filter_search}
    Retry Text Input  ${job_filter_input}  ${job}
    Retry Link Click  //clr-dg-row[contains(.,'${job}')]//a

Find Registry And Click Edit Button
    [Arguments]    ${name}
    Filter Object    ${name}
    Retry Select Object    ${name}
    Retry Element Click    ${registry_edit_btn}

Switch To Replication Manage Page
    Switch To Registries
    Switch To Replication Manage

Click Edit Button
    Retry Element Click    ${replication_rule_action}
    Retry Element Click    ${replication_rule_action_bar_edit}

Click Delete Button
    Retry Element Click    ${replication_rule_action}
    Retry Element Click    ${replication_rule_action_bar_delete}

Edit Replication Rule
    [Arguments]    ${name}
    Switch To Replication Manage Page
    Filter Replication Rule  ${name}
    Select Rule  ${name}
    Click Edit Button
    Retry Wait Until Page Contains  Edit Replication Rule

Delete Replication Rule
    [Arguments]  ${name}
    Switch To Replication Manage Page
    Filter Replication Rule  ${name}
    Select Rule  ${name}
    Click Delete Button
    Wait Until Page Contains Element  ${dialog_delete}
    Retry Double Keywords When Error  Retry Element Click  ${dialog_delete}  Retry Wait Until Page Not Contains Element  ${dialog_delete}
    Reload Page
    Filter Replication Rule  ${name}  exist=${false}

Rename Endpoint
    [arguments]  ${name}  ${newname}
    Find Registry And Click Edit Button  ${name}
    Retry Wait Until Page Contains Element  ${destination_name_xpath}
    Retry Text Input  ${destination_name_xpath}  ${newname}
    Retry Element Click  ${replication_save_xpath}

Delete Endpoint
    [Arguments]  ${name}
    Filter Registry  ${name}
    Retry Double Keywords When Error  Retry Element Click  //clr-dg-row[contains(.,'${name}')]//div[contains(@class,'clr-checkbox-wrapper')]  Checkbox Should Be Selected  //clr-dg-row[contains(.,'${name}')]//input
    Retry Element Click  ${registry_del_btn}
    Wait Until Page Contains Element  ${dialog_delete}
    Retry Element Click  ${dialog_delete}

Select Rule And Replicate
    [Arguments]  ${rule_name}
    Select Rule  ${rule_name}
    Retry Element Click    ${replication_rule_exec_id}
    Retry Double Keywords When Error    Retry Element Click    xpath=${dialog_replicate}    Retry Wait Until Page Not Contains Element    xpath=${dialog_replicate}

Image Should Be Replicated To Project
    [Arguments]  ${project}  ${image}  ${times}=6  ${tag}=${EMPTY}  ${expected_image_size_in_regexp}=${null}  ${total_artifact_count}=${null}  ${archive_count}=${null}
    FOR  ${n}  IN RANGE  0  ${times}
        Go Into Project  ${project}
        Switch To Project Repo
        ${out}  Run Keyword And Ignore Error  Retry Wait Until Page Contains  ${project}/${image}
        Log To Console  Return value is ${out[0]}
        Continue For Loop If  '${out[0]}'=='FAIL'
        Go Into Repo  ${project}  ${image}
        Wait Until Page Contains Element  //clr-dg-row[contains(., '${tag}')]//clr-dg-cell[4]/div
        ${size}=  Run Keyword If  '${tag}'!='${EMPTY}' and '${expected_image_size_in_regexp}'!='${null}'  Get Text  //clr-dg-row[contains(., '${tag}')]//clr-dg-cell[4]/div
        Run Keyword If  '${tag}'!='${EMPTY}' and '${expected_image_size_in_regexp}'!='${null}'  Should Match Regexp  '${size}'  '${expected_image_size_in_regexp}'
        ${index_out}  Go Into Index And Contain Artifacts  ${tag}  total_artifact_count=${total_artifact_count}  archive_count=${archive_count}  return_immediately=${true}
        Log All  index_out: ${index_out}
        Run Keyword If  '${index_out}'=='PASS'  Exit For Loop
        Sleep  10
    END

Verify Artifacts Counts In Archive
    [Arguments]  ${total_artifact_count}  ${tag}  ${total_artifact_count}  ${archive_count}

Executions Result Count Should Be
    [Arguments]  ${expected_status}  ${expected_trigger_type}  ${expected_result_count}
    Sleep  10
    ${count}=  Get Element Count  xpath=//clr-dg-row[contains(.,'${expected_status}') and contains(.,'${expected_trigger_type}')]
    Should Be Equal As Integers  ${count}  ${expected_result_count}

Check Latest Replication Job Status
    [Arguments]  ${expected_status}
    Retry Wait Element  //hbr-replication//div[contains(@class,'datagrid')]//clr-dg-row[1][contains(.,'${expected_status}')]

Check Latest Replication Enabled Copy By Chunk
    Retry Link Click  //hbr-replication//div[contains(@class,'datagrid')]//clr-dg-row[1]//a
    Retry Link Click  //clr-dg-row[1]//a
    Switch Window  locator=NEW
    Retry Wait Until Page Contains  chunk completed
    Switch Window  locator=MAIN
