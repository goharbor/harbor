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
Switch To Job Queues
    Retry Double Keywords When Error  Retry Element Click  //clr-main-container//clr-vertical-nav-group//span[contains(.,'Job Service Dashboard')]  Retry Wait Until Page Contains Element  ${job_service_stop_btn}

Select Jobs
    [Arguments]  @{job_types}
    FOR  ${job_type}  IN  @{job_types}
        Retry Double Keywords When Error  Retry Element Click  //clr-datagrid//clr-dg-row[contains(.,'${job_type}')]//div[contains(@class,'clr-checkbox-wrapper')]  Retry Checkbox Should Be Selected  //clr-datagrid//clr-dg-row[contains(.,'${job_type}')]//input
    END

Stop Pending Jobs
    [Arguments]  @{job_types}
    Select Jobs  @{job_types}
    Retry Double Keywords When Error  Retry Button Click  ${job_service_stop_btn}  Retry Button Click  ${confirm_btn}
    Retry Wait Until Page Contains  Stopped jobs successfully

Stop All Pending Jobs
    Retry Double Keywords When Error  Retry Button Click  ${job_service_stop_all_btn}  Retry Button Click  ${confirm_btn}
    Retry Wait Until Page Contains  Stopped all the job queues successfully

Pause Jobs
    [Arguments]  @{job_types}
    Select Jobs  @{job_types}
    Retry Double Keywords When Error  Retry Button Click  ${job_service_pause_btn}  Retry Button Click  ${confirm_btn}
    Retry Wait Until Page Contains  Paused jobs successfully
    Check Jobs Paused  Yes  @{job_types}

Resume Jobs
    [Arguments]  @{job_types}
    Select Jobs  @{job_types}
    Retry Double Keywords When Error  Retry Button Click  ${job_service_resume_btn}  Retry Button Click  ${confirm_btn}
    Retry Wait Until Page Contains  Resumed jobs successfully
    Check Jobs Paused  No  @{job_types}

Check Jobs Paused
    [Arguments]  ${paused}=No  @{job_types}
    FOR  ${job_type}  IN  @{job_types}
        Retry Double Keywords When Error  Retry Element Click  ${job_service_refresh_btn}  Retry Wait Element Visible  //clr-datagrid//clr-dg-row[contains(.,'${job_type}')]//clr-dg-cell[4][contains(.,'${paused}')]
        Select Jobs  ${job_type}
        Run Keyword If  '${paused}' == 'No'  Run Keywords
        ...  Retry Wait Element Should Be Disabled  ${job_service_resume_btn}
        ...  AND   Retry Wait Element  ${job_service_pause_btn}
        ...  ELSE  Run Keywords
        ...  Retry Wait Element Should Be Disabled  ${job_service_pause_btn}
        ...  AND   Retry Wait Element  ${job_service_resume_btn}
    END

Check Jobs Pending Count
    [Arguments]  &{jobs_type_pending_count}
    FOR  ${job_type}  IN  @{jobs_type_pending_count.keys()}
        Retry Wait Element Visible  //clr-datagrid//clr-dg-row[contains(.,'${job_type}')]//clr-dg-cell[2][contains(.,'${jobs_type_pending_count['${job_type}']}')]
    END

Check Pending Job Card
    [Arguments]  &{jobs_type_pending_count}
    ${total}=  Set Variable  ${jobs_type_pending_count['Total']}
    Remove From Dictionary  ${jobs_type_pending_count}  Total
    ${total_pending_count}=  Set Variable  0
    ${index}=  Set Variable  1
    FOR  ${job_type}  IN  @{jobs_type_pending_count.keys()}
        Run Keyword If  '${total}' != '0'  Retry Wait Until Page Contains Element  //app-pending-job-card//div[contains(@class,'clr-row ng-star-inserted')][${index}]//div[1][contains(.,'${job_type}')]
        Retry Wait Until Page Contains Element  //app-pending-job-card//div[contains(@class,'clr-row ng-star-inserted')][${index}]//div[2][contains(.,'${jobs_type_pending_count['${job_type}']}')]
        ${index}=  Evaluate  ${index} + 1
        ${total_pending_count}=  Evaluate  ${total_pending_count} + ${jobs_type_pending_count['${job_type}']}
    END
    Retry Wait Element Visible  //app-pending-job-card//div[contains(text(),'Total: ${total_pending_count}')]

Check Jobs Latency
    [Arguments]  &{jobs_type_is_zore}
    FOR  ${job_type}  IN  @{jobs_type_is_zore.keys()}
        ${latency_xpath}=  Set Variable  //clr-datagrid//clr-dg-row[contains(.,'${job_type}')]//clr-dg-cell[3]//span[text()='0']
        Run Keyword If  ${jobs_type_is_zore['${job_type}']}==${true}  Retry Wait Until Page Contains Element  ${latency_xpath}
        ...  ELSE  Retry Wait Until Page Not Contains Element  ${latency_xpath}
    END

Check Button Status
    # Pause is Yes
    Select Jobs  GARBAGE_COLLECTION
    Retry Wait Element  ${job_service_resume_btn}
    Retry Wait Element Should Be Disabled  ${job_service_pause_btn}
    # Pause is Yes
    Select Jobs  PURGE_AUDIT_LOG
    Retry Wait Element  ${job_service_resume_btn}
    Retry Wait Element Should Be Disabled  ${job_service_pause_btn}
    # Pause is No
    Select Jobs  SCAN_DATA_EXPORT
    Retry Wait Element Should Be Disabled  ${job_service_resume_btn}
    Retry Wait Element Should Be Disabled  ${job_service_pause_btn}
    # Refresh
    Retry Element Click  ${job_service_refresh_btn}
    # Pause is No
    Select Jobs  REPLICATION
    Retry Wait Element  ${job_service_pause_btn}
    Retry Wait Element Should Be Disabled  ${job_service_resume_btn}
    # Pause is No
    Select Jobs  P2P_PREHEAT
    Retry Wait Element  ${job_service_pause_btn}
    Retry Wait Element Should Be Disabled  ${job_service_resume_btn}
    # Pause is Yes
    Select Jobs  IMAGE_SCAN
    Retry Wait Element Should Be Disabled  ${job_service_resume_btn}
    Retry Wait Element Should Be Disabled  ${job_service_pause_btn}
