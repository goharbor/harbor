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
${HARBOR_VERSION}  v1.1.1

*** Keywords ***
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
    Run Keyword If  '${provider}' != 'dockerHub'  Run keyword  Retry Text Input  xpath=${destination_url_xpath}  ${url}
    Retry Text Input  xpath=${destination_username_xpath}  ${username}
    Retry Text Input  xpath=${destination_password_xpath}  ${pwd}
    #cancel verify cert since we use a selfsigned cert
    Retry Element Click  ${destination_insecure_xpath}
    Run Keyword If  '${save}' == 'Y'  Run keyword  Retry Double Keywords When Error  Retry Element Click  ${replication_save_xpath}  Retry Wait Until Page Not Contains Element  ${replication_save_xpath}
    Run Keyword If  '${save}' == 'N'  No Operation

Create A Rule With Existing Endpoint
# day 1=Monday..7=Sunday timeformat 12hour+am/pm
    [Arguments]  ${name}  ${project_name}  ${endpoint}  ${mode}  ${plan}=Daily  ${weekday}=1  ${time}=0800a
    #click new
    Retry Element Click  ${new_name_xpath}
    #input name
    Input Text  ${rule_name}  ${name}
    #input descripiton,here skip, leave it blank
    #source projects, input
    Input Text  ${source_project}  ${project_name}
    #set filter
    Retry Element Click  ${source_image_filter_add}
    Input Text  ${source_iamge_repo_filter}  *
    Retry Element Click  ${source_image_filter_add}
    Input Text  ${source_image_tag_filter}  *
    #select endpoint
    Retry Element Click  ${rule_target_select}
    Wait Until Element Is Visible  //select[@id='ruleTarget']//option[contains(.,'${endpoint}')]
    Retry Element Click  //select[@id='ruleTarget']//option[contains(.,'${endpoint}')]
    #set trigger
    Retry Element Click  ${rule_trigger_select}
    Wait Until Element Is Visible  //select[@id='ruleTrigger']//option[contains(.,'${mode}')]
    Retry Element Click  //select[@id='ruleTrigger']//option[contains(.,'${mode}')]
    Run Keyword If  '${mode}' == 'Scheduled'  Setting Replicaiton Schedule  ${plan}  ${weekday}  ${time}
    #click save
    Retry Element Click  ${rule_save_button}

Project Create A Rule With Existing Endpoint
# day 1=Monday..7=Sunday timeformat 12hour+am/pm
    [Arguments]  ${name}  ${project_name}  ${endpoint}  ${mode}  ${plan}=Daily  ${weekday}=1  ${time}=0800a
    #click new
    Retry Element Click  ${new_name_xpath}
    #input name
    Input Text  ${rule_name}  ${name}
    #input descripiton,here skip, leave it blank
    #in this keyword, source project is not need to input
    #set filter
    Retry Element Click  ${source_image_filter_add}
    Input Text  ${source_iamge_repo_filter}  *
    Retry Element Click  ${source_image_filter_add}
    Input Text  ${source_image_tag_filter}  *
    #select endpoint
    Retry Element Click  ${rule_target_select}
    Wait Until Element Is Visible  //select[@id='ruleTarget']//option[contains(.,'${endpoint}')]
    Retry Element Click  //select[@id='ruleTarget']//option[contains(.,'${endpoint}')]
    #set trigger  
    Retry Element Click  ${rule_trigger_select}
    Wait Until Element Is Visible  //select[@id='ruleTrigger']//option[contains(.,'${mode}')]
    Retry Element Click  //select[@id='ruleTrigger']//option[contains(.,'${mode}')]
    Run Keyword If  '${mode}' == 'Scheduled'  Setting Replicaiton Schedule  ${plan}  ${weekday}  ${time}
    #click save
    Retry Element Click  ${rule_save_button}

Setting Replication Schedule
    [Arguments]  ${plan}  ${weekday}=1  ${time}=0800a
    Retry Element Click  ${schedule_type_select}
    Wait Until Element Is Visible  //select[@name='scheduleType']/option[@value='${plan}']
    Retry Element Click  //select[@name='scheduleType']/option[@value='${plan}']
    Run Keyword If  '${plan}' == 'Weekly'  Setting Replication Weekday  ${weekday}
    Input Text  ${shcedule_time}  ${time}

Setting Replication Weekday
    [arguments]  ${day}
    Retry Element Click  ${schedule_day_select}
    Wait Until Element Is Visible  //select[@name='scheduleDay']/option[@value='${day}']
    Retry Element Click  //select[@name='scheduleDay']/option[@value='${day}']

Endpoint Is Unpingable
    Retry Element Click  ${ping_test_button}
    Wait Until Page Contains  Failed

Endpoint Is Pingable
    Retry Element Click  ${ping_test_button}
    Wait Until Page Contains  successfully

Disable Certificate Verification
    Checkbox Should Be Selected  ${destination_insecure_checkbox}
    Retry Element Click  ${destination_insecure_xpath}
    Sleep  1

Enable Certificate Verification
    Checkbox Should Not Be Selected  ${destination_insecure_checkbox}
    Retry Element Click  ${destination_insecure_xpath}
    Sleep  1

Switch To Registries
    Retry Element Click  ${nav_to_registries}
    Sleep  1

Switch To Replication Manage
    Retry Element Click  ${nav_to_replications}
    Sleep  1

Trigger Replication Manual
    [Arguments]  ${rule}
    Retry Element Click  ${rule_filter_search}
    Input Text   ${rule_filter_input}  ${rule}
    Sleep  1
    Retry Element Click  //clr-dg-row[contains(.,'${rule}')]//label
    Retry Element Click  ${action_bar_replicate}
    Wait Until Page Contains Element  ${dialog_replicate}
    #change from click to mouse down and up 
    Mouse Down  ${dialog_replicate}
    Mouse Up  ${dialog_replicate}
    Sleep  2
    Wait Until Page Contains Element  //clr-tab-content//div[contains(.,'${rule}')]/../div/clr-icon[@shape='success-standard']
    Sleep  1

Rename Rule
    [Arguments]  ${rule}  ${newname}
    Retry Element Click  ${rule_filter_search}
    Input Text  ${rule_filter_input}  ${rule}
    Sleep  1
    Retry Element Click  //clr-dg-row[contains(.,'${rule}')]//label
    Retry Element Click  ${action_bar_edit}
    Input Text  ${rule_name}  ${newname}
    Retry Element Click  ${rule_save_button}

Delete Rule
    [Arguments]  ${rule} 
    Retry Element Click  ${rule_filter_search}
    Input Text   ${rule_filter_input}  ${rule}
    Sleep  1
    Retry Element Click  //clr-dg-row[contains(.,'${rule}')]//label
    Retry Element Click  ${action_bar_delete}
    Wait Until Page Contains Element  ${dialog_delete}
    #change from click to mouse down and up
    Mouse Down  ${dialog_delete}
    Mouse Up  ${dialog_delete}
    Sleep  2

Filter Rule
    [Arguments]  ${rule} 
    Retry Element Click  ${rule_filter_search}
    Input Text   ${rule_filter_input}  ${rule}
    Sleep  1

Select Rule
    [Arguments]  ${rule}
    Sleep  1
    Retry Element Click  //clr-dg-row[contains(.,'${rule}')]//label

Stop Jobs
    Retry Element Click  ${stop_jobs_button}

View Job Log
    [arguments]  ${job}
    Retry Element Click  ${job_filter_search}
    Input Text  ${job_filter_input}  ${job}
    Click Link  //clr-dg-row[contains(.,'${job}')]//a

Rename Endpoint
    [arguments]  ${name}  ${newname}
    Filter Object  ${name}
    Select Object  ${name}
    Retry Element Click  ${action_bar_edit}
    Wait Until Page Contains Element  ${destination_name_xpath}
    Input Text  ${destination_name_xpath}  ${newname}
    Retry Element Click  ${replication_save_xpath}

Delete Endpoint
    [Arguments]  ${name}
    Retry Element Click  ${endpoint_filter_search}
    Input Text   ${endpoint_filter_input}  ${name}
    Sleep  1
    #click checkbox before target endpoint
    Retry Element Click  //clr-dg-row[contains(.,'${name}')]//label
    Sleep  1
    Retry Element Click  ${action_bar_delete}
    Wait Until Page Contains Element  ${dialog_delete}
    Retry Element Click  ${dialog_delete}
