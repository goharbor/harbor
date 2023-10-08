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

*** Keywords ***
Switch To Log Rotation
    Retry Element Click  //clr-main-container//clr-vertical-nav-group//span[contains(.,'Clean Up')]
    Retry Element Click  ${log_rotation_page_xpath}

Purge Now
    [Arguments]  ${keep_records}  ${keep_records_unit}  ${expected_status}=SUCCESS  ${exclude_operations}=@{EMPTY}
    Retry Text Input  ${keep_records_input}  ${keep_records}
    Retry Double Keywords When Error  Retry Element Click  ${keep_records_unit_select}  Retry Element Click  ${keep_records_unit_select}//option[contains(.,'${keep_records_unit}')]
    ${len}=  Get Length  ${exclude_operations}
    Run Keyword If  ${len} > 0  Click Exclude Operation  @{exclude_operations}
    Retry Double Keywords When Error  Retry Button Click  ${purge_now_btn}  Retry Wait Until Page Contains Element  ${latest_purge_job_status_xpath}\[contains(.,'${expected_status}')]
    Run Keyword If  '${expected_status}' == 'SUCCESS'  Retry Action Keyword  Verify Last completed Time

Click Exclude Operation
    [Arguments]  @{exclude_operations}
    FOR  ${element}  IN  @{exclude_operations}
        Retry Element Click  //form//div//label[contains(@class,'clr-control-label') and contains(.,'${element}')]
    END

Verify Last completed Time
    ${latest_purge_job_update_time}=  Get Text  ${latest_purge_job_update_time_xpath}
    ${purge_job_last_completed_time}=  Get Text  ${purge_job_last_completed_time_xpath}
    Should Be Equal  ${latest_purge_job_update_time}  ${purge_job_last_completed_time}

Set Log Rotation Schedule
    [Arguments]  ${keep_records}  ${keep_records_unit}  ${type}  ${cron}=${null}  ${exclude_operations}=@{EMPTY}
    Retry Text Input  ${keep_records_input}  ${keep_records}
    Retry Double Keywords When Error  Retry Element Click  ${keep_records_unit_select}  Retry Element Click  ${keep_records_unit_select}//option[contains(.,'${keep_records_unit}')]
    Retry Button Click  ${log_rotation_schedule_edit_btn}
    Retry Double Keywords When Error  Retry Element Click  ${log_rotation_schedule_select}  Retry Element Click  ${log_rotation_schedule_select}//option[contains(.,'${type}')]
    Run Keyword If  '${type}' == 'Custom'  Retry Text Input  ${log_rotation_schedule_cron_input}  ${cron}
    ${len}=  Get Length  ${exclude_operations}
    Run Keyword If  ${len} > 0  Click Exclude Operation  @{exclude_operations}
    Retry Double Keywords When Error  Retry Button Click  ${log_rotation_schedule_save_btn}  Retry Wait Until Page Contains  Purge schedule has been reset

Get Purge Job Results
    ${cmd}=  Set Variable  curl -u ${HARBOR_ADMIN}:${HARBOR_PASSWORD} -s --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/v2.0/system/purgeaudit?sort=-creation_time&page=1&page_size=100"
    ${rc}  ${output}=  Run And Return Rc And Output  ${cmd}
    Log  ${output}
    [Return]  ${output}
