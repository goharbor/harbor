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
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${ip}
${SSH_USER}  root
${HARBOR_ADMIN}  admin

*** Test Cases ***
# Due to Docker 20's new behavior, let 'Proxy Cache' be the 1st case to run
#   and at the same time all images to be pull among all cases should be not exsit before pulling.
Test Case - Proxy Cache
    [Tags]  proxy_cache
    ${d}=  Get Current Date    result_format=%m%s
    ${registry}=  Set Variable  https://${LOCAL_REGISTRY}
    ${user_namespace}=  Set Variable  nightly
    ${image}=  Set Variable  for_proxy
    ${tag}=  Set Variable  1.0
    ${manifest_index}=  Set Variable  alpine
    ${manifest_tag}=  Set Variable  3.12.0
    ${test_user}=  Set Variable  user010
    ${test_pwd}=  Set Variable  Test1@34
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Registries
    Create A New Endpoint  harbor  e1${d}  ${registry}  ${null}  ${null}
    Create An New Project And Go Into Project  project${d}  proxy_cache=${true}  registry=e1${d}
    Manage Project Member Without Sign In  project${d}  ${test_user}  Add  has_image=${false}
    Go Into Project  project${d}  has_image=${false}
    Change Member Role  ${test_user}  Developer
    Pull Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${user_namespace}/${image}  tag=${tag}
    Pull Image  ${ip}  ${test_user}  ${test_pwd}  project${d}  ${user_namespace}/${manifest_index}  tag=${manifest_tag}
    Log To Console  Start to Sleep 3 minitues......
    Sleep  180
    Go Into Repo  project${d}  ${user_namespace}/${image}

    FOR  ${idx}  IN RANGE  0  15
        Log All  Checking manifest ${idx} round......
        Sleep  60
        Go Into Project  project${d}
        ${repo_out}=  Run Keyword And Ignore Error  Go Into Repo  project${d}  ${user_namespace}/${manifest_index}
        Continue For Loop If  '${repo_out[0]}'=='FAIL'
        ${artifact_out}=  Run Keyword And Ignore Error  Go Into Index And Contain Artifacts  ${manifest_tag}  total_artifact_count=1
        Exit For Loop If  '${artifact_out[0]}'=='PASS'
    END
    Should Be Equal As Strings  '${artifact_out[0]}'  'PASS'

    Cannot Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  busybox:latest  err_msg=can not push artifact to a proxy project
    Cannot Push image  ${ip}  ${test_user}  ${test_pwd}  project${d}  busybox:latest  err_msg=can not push artifact to a proxy project
    Close Browser

Test Case - GC Schedule Job
    [tags]  GC_schedule
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%M
    Log To Console  GC Schedule Job ${d}
    ${project_name}=  Set Variable  gc_schedule_proj${d}
    ${image}=  Set Variable  redis
    ${sha256}=  Set Variable  e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  ${project_name}
    Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image}  sha256=${sha256}
    Sleep  50
    Go Into Repo  ${project_name}  ${image}
    Switch To Garbage Collection
    Set GC Schedule  custom  value=0 */2 * * * *
    Sleep  480
    Set GC Schedule  none
    Sleep  180
    ${logs}=  Get GC Logs
    ${logs}=  Should Match Regexp  ${logs}  \\\[(.+)\\\]
    Log All  logs:${logs}[1]
    ${logs} = 	Replace String 	${logs}[1] 	\\ 	${EMPTY} 	count=-1
    ${logs} = 	Replace String 	${logs} 	"{ 	{ 	count=-1
    ${logs} = 	Replace String 	${logs} 	}" 	} 	count=-1
    Log All  str:${logs}
    ${logs_list}=  Get Regexp Matches  ${logs}  {"creation_time.+?\\d{3}Z"}
    Log All  logs_list:${logs_list}
    ${len}=  Get Length  ${logs_list}
    Log All  len:${len}
    FOR  ${log}  IN  @{logs_list}
        Log All  log:${log}
        ${log_json}=  evaluate  json.loads('''${log}''')
        Log All  log_json:${log_json}
        Should Be Equal As Strings  ${log_json["job_kind"]}  SCHEDULE
        Should Be Equal As Strings  '${log_json["job_name"]}'  'GARBAGE_COLLECTION'
    END
    #Only return latest 10 records for GC job
    Should Be True  ${len} > 3 and ${len} < 6
    Close Browser

Test Case - Scan Schedule Job
    [tags]  Scan_schedule
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%M
    Log To Console  ${d}
    ${project_name}=  Set Variable  scan_schedule_proj${d}
    ${image}=  Set Variable  redis
    ${sha256}=  Set Variable  e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  ${project_name}
    Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image}  sha256=${sha256}
    Sleep  50
    Go Into Repo  ${project_name}  ${image}
    Retry Wait Until Page Contains Element  ${not_scanned_icon}
    Switch To Vulnerability Page
    ${flag}=  Set Variable  ${false}
    FOR    ${i}    IN RANGE    999999
        ${minite}=  Get Current Date  result_format=%M
        ${minite_int} =  Convert To Integer  ${minite}
        ${left} =  Evaluate 	${minite_int}%10
        Log To Console    ${i}/${left}
        Sleep  55
        Run Keyword If  ${left} <= 3 and ${left} != 0   Run Keywords  Set Scan Schedule  Custom  value=0 */10 * * * *  AND  Set Suite Variable  ${flag}  ${true}
        Exit For Loop If    '${flag}' == '${true}'
    END
    # After scan custom schedule is set, image should stay in unscanned status.
    Log To Console  Sleep for 300 seconds......
    Sleep  180
    Go Into Repo  ${project_name}  ${image}
    Retry Wait Until Page Contains Element  ${not_scanned_icon}

    Log To Console  Sleep for 500 seconds......
    Sleep  500
    Go Into Repo  ${project_name}  ${image}
    Scan Result Should Display In List Row  ${sha256}
    View Repo Scan Details  Critical  High
    Close Browser

Test Case - Replication Schedule Job
    [tags]  Replication_schedule
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%M
    Log To Console  ${d}
    ${project_name}=  Set Variable  replication_schedule_proj${d}
    ${image_a}=  Set Variable  mariadb
    ${tag_a}=  Set Variable  111
    ${image_b}=  Set Variable  centos
    ${tag_b}=  Set Variable  222
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  ${project_name}
    Switch To Registries
    Create A New Endpoint    harbor    e${d}    https://${LOCAL_REGISTRY}    ${null}    ${null}    Y
    Switch To Replication Manage
    ${flag}=  Set Variable  ${false}
    FOR    ${i}    IN RANGE    999999
        ${minite}=  Get Current Date  result_format=%M
        ${minite_int} =  Convert To Integer  ${minite}
        ${left} =  Evaluate 	${minite_int}%10
        Log To Console    ${i}/${left}
        Run Keyword If  ${left} <= 3 and ${left} != 0   Run Keywords  Create A Rule With Existing Endpoint    rule${d}    pull    nightly/{mariadb,centos}    image    e${d}    ${project_name}  mode=Scheduled  cron=0 */10 * * * *  AND  Set Suite Variable  ${flag}  ${true}
        Sleep  40
        Exit For Loop If    '${flag}' == '${true}'
    END

    # After replication schedule is set, project should contain 2 images.
    Log To Console  Sleep for 720 seconds......
    Sleep  720
    Go Into Repo  ${project_name}  ${image_a}
    Artifact Exist  ${tag_a}
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}  ${image_b}
    Artifact Exist  ${tag_b}

    # Delete repository
    Go Into Project  ${project_name}
    Delete Repo  ${project_name}  ${image_a}
    Delete Repo  ${project_name}  ${image_b}

    # After replication schedule is set, project should contain 2 images.
    Log To Console  Sleep for 600 seconds......
    Sleep  600
    Go Into Repo  ${project_name}  ${image_a}
    Artifact Exist  ${tag_a}
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}  ${image_b}
    Artifact Exist  ${tag_b}
    Close Browser

Test Case - P2P Preheat Schedule Job
    [Tags]  p2p_preheat_schedule  need_distribution_endpoint
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%M
    ${project_name}=  Set Variable  p2p_preheat_schedule_proj${d}
    ${dist_name}=  Set Variable  distribution${d}
    ${policy_name}=  Set Variable  policy${d}
    ${image}=  Set Variable  busybox
    ${tag}=  Set Variable  latest
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Distribution  Dragonfly  ${dist_name}  ${DISTRIBUTION_ENDPOINT}
    Create An New Project And Go Into Project  ${project_name}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image}  ${tag}
    Create An New P2P Preheat Policy  ${policy_name}  ${dist_name}  **  **
    Set P2P Preheat Policy Schedule  ${policy_name}  Custom  0 */2 * * * *
    Sleep  480
    Edit A P2P Preheat Policy  ${policy_name}  **  Manual
    Sleep  180
    ${logs}=  Get P2P Preheat Logs  ${project_name}  ${policy_name}
    ${logs}=  Evaluate  json.loads("""${logs}""")  json
    ${len}=  Get Length  ${logs}
    FOR  ${log}  IN  @{logs}
        Log  ${log}
        Should Be Equal As Strings  ${log["trigger"]}  scheduled
        Should Be Equal As Strings  ${log["status"]}  Success
        Should Be Equal As Strings  ${log["vendor_type"]}  P2P_PREHEAT
    END
    Should Be True  ${len} > 3 and ${len} < 6
    Delete A P2P Preheat Policy  ${policy_name}
    Delete A Distribution  ${dist_name}  ${DISTRIBUTION_ENDPOINT}
    Close Browser

Test Case - Log Rotation Schedule Job
    [Tags]  log_rotation_schedule
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Log Rotation
    ${exclude_operations}  Create List  Pull
    Set Log Rotation Schedule  2  Days  Custom  0 */2 * * * *  ${exclude_operations}
    Sleep  480
    Set Log Rotation Schedule  2  Days  None
    Sleep  180
    ${logs}=  Get Purge Job Results
    ${logs}=  Replace String  ${logs}  "{  {  count=-1
    ${logs}=  Replace String  ${logs}  }"  }  count=-1
    ${logs}=  Evaluate  json.loads("""${logs}""")  json
    ${len}=  Get Length  ${logs}
    FOR  ${log}  IN  @{logs}
        Log  ${log}
        Should Be Equal As Strings  ${log["job_name"]}  PURGE_AUDIT_LOG
        Should Be Equal As Strings  ${log["job_kind"]}  SCHEDULE
        Should Be Equal As Strings  ${log["job_status"]}  Success
        Should Be Equal As Strings  ${log["job_parameters"]["audit_retention_hour"]}  48
        Should Be Equal As Strings  ${log["job_parameters"]["dry_run"]}  False
        Should Not Contain Any  ${log["job_parameters"]["include_operations"]}  @{exclude_operations}  ignore_case=True
    END
    Should Be True  ${len} > 3 and ${len} < 6
    Close Browser

Test Case - Job Service Dashboard Schedules
    [Tags]  job_service_schedules
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${schedule_type}=  Set Variable  Custom
    ${schedule_cron}=  Set Variable  0 */2 * * * *
    ${image}=  Set Variable  photon
    ${tag}=  Set Variable  2.0
    ${project_name}=  Set Variable  project${d}
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  ${project_name}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image}  ${tag}  ${tag}
    ${replication_policy_name}  ${p2p_policy_name}  ${distribution_name}=  Create Schedules For Job Service Dashboard Schedules  ${project_name}  ${schedule_type}  ${schedule_cron}  ${DISTRIBUTION_ENDPOINT}
    Switch To Job Schedules
    Check Schedule List  ${schedule_cron}
    Sleep  150
    Pause All Schedules
    # Check schedule is running
    Go Into Project  ${project_name}
    # Check that the tag tetention schedule is running
    Switch To Tag Retention
    Wait Until Element Is Visible And Enabled  //tag-retention//clr-datagrid//clr-dg-row[1]//clr-dg-cell[5]
    ${tag_retention_start_time1}=  Get Text  //tag-retention//clr-datagrid//clr-dg-row[1]//clr-dg-cell[5]
    # Check that the preheat policy schedule is running
    Switch To P2P Preheat
    Select P2P Preheat Policy  ${p2p_policy_name}
    Wait Until Element Is Visible And Enabled  //div[./h4[text()='Executions']]//clr-datagrid//clr-dg-row[1]//clr-dg-cell[4]
    ${preheat_start_time1}=  Get Text  //div[./h4[text()='Executions']]//clr-datagrid//clr-dg-row[1]//clr-dg-cell[4]
    # Check that the replication schedule is running
    Switch To Replication Manage
    Select Rule  ${replication_policy_name}
    Wait Until Element Is Visible And Enabled  //clr-datagrid[.//span[text()='Total']]//clr-dg-row[1]//clr-dg-cell[4]
    ${replication_start_time1}=  Get Text  //clr-datagrid[.//span[text()='Total']]//clr-dg-row[1]//clr-dg-cell[4]
    # Check that the scan all schedule is running
    ${artifact_info}=  Get The Specific Artifact  ${project_name}  ${image}  ${tag}
    ${artifact_info}=  Evaluate  json.loads("""${artifact_info}""")  json
    ${scan_all_start_time1}=  Set Variable  ${artifact_info["scan_overview"]["application/vnd.security.vulnerability.report; version=1.1"]["start_time"]}
    # Check that the GC schedule is running
    Switch To Garbage Collection
    Wait Until Element Is Visible And Enabled  //clr-datagrid//clr-dg-row[1]//clr-dg-cell[5]
    ${gc_start_time1}=  Get Text  //clr-datagrid//clr-dg-row[1]//clr-dg-cell[5]
    # Check that the log rotation schedule is running
    Switch To Log Rotation
    Wait Until Element Is Visible And Enabled  //clr-datagrid//clr-dg-row[1]//clr-dg-cell[5]
    ${log_rotation_start_time1}=  Get Text  //clr-datagrid//clr-dg-row[1]//clr-dg-cell[5]
    Sleep  150
    # Check schedules is paused
    Go Into Project  ${project_name}
    # Check that the tag tetention schedule is paused
    Switch To Tag Retention
    Wait Until Element Is Visible And Enabled  //tag-retention//clr-datagrid//clr-dg-row[1]//clr-dg-cell[5]
    ${tag_retention_start_time2}=  Get Text  //tag-retention//clr-datagrid//clr-dg-row[1]//clr-dg-cell[5]
    Should Be Equal  ${tag_retention_start_time1}  ${tag_retention_start_time2}
    # Check that the preheat policy schedule is paused
    Switch To P2P Preheat
    Select P2P Preheat Policy  ${p2p_policy_name}
    Wait Until Element Is Visible And Enabled  //div[./h4[text()='Executions']]//clr-datagrid//clr-dg-row[1]//clr-dg-cell[4]
    ${preheat_start_time2}=  Get Text  //div[./h4[text()='Executions']]//clr-datagrid//clr-dg-row[1]//clr-dg-cell[4]
    Should Be Equal  ${preheat_start_time1}  ${preheat_start_time2}
    # Check that the replication schedule is paused
    Switch To Replication Manage
    Select Rule  ${replication_policy_name}
    Wait Until Element Is Visible And Enabled  //clr-datagrid[.//span[text()='Total']]//clr-dg-row[1]//clr-dg-cell[4]
    ${replication_start_time2}=  Get Text  //clr-datagrid[.//span[text()='Total']]//clr-dg-row[1]//clr-dg-cell[4]
    Should Be Equal  ${replication_start_time1}  ${replication_start_time2}
    # Check that the scan all schedule is paused
    ${artifact_info}=  Get The Specific Artifact  ${project_name}  ${image}  ${tag}
    ${artifact_info}=  Evaluate  json.loads("""${artifact_info}""")  json
    ${scan_all_start_time2}=  Set Variable  ${artifact_info["scan_overview"]["application/vnd.security.vulnerability.report; version=1.1"]["start_time"]}
    Should Be Equal  ${scan_all_start_time1}  ${scan_all_start_time2}
    # Check that the GC schedule is paused
    Switch To Garbage Collection
    Wait Until Element Is Visible And Enabled  //clr-datagrid//clr-dg-row[1]//clr-dg-cell[5]
    ${gc_start_time2}=  Get Text  //clr-datagrid//clr-dg-row[1]//clr-dg-cell[5]
    Should Be Equal  ${gc_start_time1}  ${gc_start_time2}
    # Check that the log rotation schedule is paused
    Switch To Log Rotation
    Wait Until Element Is Visible And Enabled  //clr-datagrid//clr-dg-row[1]//clr-dg-cell[5]
    ${log_rotation_start_time2}=  Get Text  //clr-datagrid//clr-dg-row[1]//clr-dg-cell[5]
    Should Be Equal  ${log_rotation_start_time1}  ${log_rotation_start_time2}
    Reset Schedules For Job Service Dashboard Schedules  ${project_name}  ${replication_policy_name}  ${p2p_policy_name}
    Delete A Distribution  ${distribution_name}  ${DISTRIBUTION_ENDPOINT}
    Switch To Job Schedules
    Resume All Schedules
    Close Browser
