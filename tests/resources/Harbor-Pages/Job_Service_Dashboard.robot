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

Switch To Job Schedules
    Retry Element Click  //clr-main-container//clr-vertical-nav-group//span[contains(.,'Job Service Dashboard')]
    Retry Double Keywords When Error  Retry Button Click  ${job_service_schedules_btn}  Retry Wait Until Page Contains  Vendor Type

Switch To Job Workers
    Retry Element Click  //clr-main-container//clr-vertical-nav-group//span[contains(.,'Job Service Dashboard')]
    Retry Double Keywords When Error  Retry Button Click  ${job_service_workers_btn}  Retry Wait Until Page Contains  Worker Pools

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

Check Schedule List
    [Arguments]  ${schedule_cron}
    # Check retention policy schedule
    Retry Wait Until Page Contains Element  //clr-dg-row[.//clr-dg-cell[text()='RETENTION'] and .//clr-dg-cell[text()='${schedule_cron}']]
    # Check preheat policy schedule
    Retry Wait Until Page Contains Element  //clr-dg-row[.//clr-dg-cell[text()='P2P_PREHEAT'] and .//clr-dg-cell[text()='${schedule_cron}']]
    # Check replication policy schedule
    Retry Wait Until Page Contains Element  //clr-dg-row[.//clr-dg-cell[text()='REPLICATION'] and .//clr-dg-cell[text()='${schedule_cron}']]
    # Check scan all schedule
    Retry Wait Until Page Contains Element  //clr-dg-row[.//clr-dg-cell[text()='SCAN_ALL'] and .//clr-dg-cell[text()='${schedule_cron}']]
    # Check GC schedule
    Retry Wait Until Page Contains Element  //clr-dg-row[.//clr-dg-cell[text()='GARBAGE_COLLECTION'] and .//clr-dg-cell[text()='${schedule_cron}']]
    # Check log rotation schedule
    Retry Wait Until Page Contains Element  //clr-dg-row[.//clr-dg-cell[text()='PURGE_AUDIT_LOG'] and .//clr-dg-cell[text()='${schedule_cron}']]
    # Check execution sweep schedule
    Retry Wait Until Page Contains Element  //clr-dg-row[.//clr-dg-cell[text()='EXECUTION_SWEEP'] and .//clr-dg-cell[text()='0 0 0 * * *']]
    # Check system artifact cleanup schedule
    Retry Wait Until Page Contains Element  //clr-dg-row[.//clr-dg-cell[text()='SYSTEM_ARTIFACT_CLEANUP'] and .//clr-dg-cell[text()='0 0 0 * * *']]


Pause All Schedules
    Retry Double Keywords When Error  Retry Button Click  ${job_service_schedules_pause_all_btn}  Retry Button Click  ${confirm_btn}
    Retry Wait Until Page Contains  Paused all the schedules successfully
    Retry Wait Until Page Contains Element  //app-schedule-card//span[text()='Paused']

Resume All Schedules
    Retry Double Keywords When Error  Retry Button Click  ${job_service_schedules_resume_all_btn}  Retry Button Click  ${confirm_btn}
    Retry Wait Until Page Contains  Resumed all the schedules successfully
    Retry Wait Until Page Contains Element  //app-schedule-card//span[text()='Running']

Check Schedules Status Is Pause
    [Arguments]  ${project_name}  ${replication_rule_name}  ${p2p_policy_name}
    # Check that the retention policy schedule is Pause
    Go Into Project  ${project_name}
    Switch To Tag Retention
    Retry Wait Until Page Contains Element  //span[text()='Schedule has been paused']
    # Check that the preheat policy schedule is Pause
    Switch To P2P Preheat
    Retry Wait Until Page Contains Element  //clr-dg-row[.//clr-dg-cell[text()='${p2p_policy_name}'] and .//clr-dg-cell[text()=' Scheduled(Paused) ']]
    # Check that the replication policy schedule is Pause
    Switch To Replication Manage
    Retry Wait Until Page Contains Element  //clr-dg-row[.//clr-dg-cell[text()='${replication_rule_name}'] and .//clr-dg-cell[text()=' Scheduled(Paused) ']]
    # Check that the scan all schedule is Pause
    Switch To Vulnerability Page
    Retry Wait Until Page Contains Element  //span[text()='Schedule to scan all has been paused']
    # Check that the GC schedule is Pause
    Switch To Garbage Collection
    Retry Wait Until Page Contains Element  //span[text()='Schedule to GC has been paused']
    # Check that the log rotation schedule is Pause
    Switch To Log Rotation
    Retry Wait Until Page Contains Element  //span[text()='Schedule to purge has been paused']

Check Schedules Status Is Not Pause
    [Arguments]  ${project_name}  ${replication_rule_name}  ${p2p_policy_name}
     # Check that the retention policy schedule is not Pause
    Go Into Project  ${project_name}
    Switch To Tag Retention
    Retry Wait Until Page Not Contains Element  //span[text()='Schedule has been paused']
    # Check that the preheat policy schedule is not Pause
    Switch To P2P Preheat
    Retry Wait Until Page Not Contains Element  //clr-dg-row[.//clr-dg-cell[text()='${p2p_policy_name}'] and .//clr-dg-cell[text()=' Scheduled(Paused) ']]
    # Check that the replication policy schedule is not Pause
    Switch To Replication Manage
    Retry Wait Until Page Not Contains Element  //clr-dg-row[.//clr-dg-cell[text()='${replication_rule_name}'] and .//clr-dg-cell[text()=' Scheduled(Paused) ']]
    # Check that the scan all schedule is not Pause
    Switch To Vulnerability Page
    Retry Wait Until Page Not Contains Element  //span[text()='Schedule to scan all has been paused']
    # Check that the GC schedule is not Pause
    Switch To Garbage Collection
    Retry Wait Until Page Not Contains Element  //span[text()='Schedule to GC has been paused']
    # Check that the log rotation schedule is not Pause
    Switch To Log Rotation
    Retry Wait Until Page Not Contains Element  //span[text()='Schedule to purge has been paused']

Check Worker Log
    [Arguments]  ${job_name}  ${expected_log}
    Retry Link Click  //clr-datagrid[.//button[text()='Worker ID']]//clr-dg-row[.//clr-dg-cell[text()='${job_name}']]//a
    Switch Window  locator=NEW
    Retry Wait Until Page Contains  ${expected_log}
    Switch Window  locator=MAIN
