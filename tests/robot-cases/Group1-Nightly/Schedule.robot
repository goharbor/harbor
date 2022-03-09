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
    ${registry}=  Set Variable  https://cicd.harbor.vmwarecna.net
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
    Go Into Project  project${d}
    Go Into Repo  project${d}/${user_namespace}/${image}

    FOR  ${idx}  IN RANGE  0  15
        Log All  Checking manifest ${idx} round......
        Sleep  60
        Go Into Project  project${d}
        ${repo_out}=  Run Keyword And Ignore Error  Go Into Repo  project${d}/${user_namespace}/${manifest_index}
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
    Go Into Repo  ${project_name}/${image}
    Switch To Garbage Collection
    Switch To GC History
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
    Go Into Repo  ${project_name}/${image}
    Retry Wait Until Page Contains Element  ${not_scanned_icon}
    Switch To Vulnerability Page
    ${flag}=  Set Variable  ${false}
    FOR    ${i}    IN RANGE    999999
        ${minite}=  Get Current Date  result_format=%M
        ${minite_int} =  Convert To Integer  ${minite}
        ${left} =  Evaluate 	${minite_int}%10
        Log To Console    ${i}/${left}
        Sleep  55
        Run Keyword If  ${left} <= 3 and ${left} != 0   Run Keywords  Set Scan Schedule  custom  value=0 */10 * * * *  AND  Set Suite Variable  ${flag}  ${true}
        Exit For Loop If    '${flag}' == '${true}'
    END
    # After scan custom schedule is set, image should stay in unscanned status.
    Log To Console  Sleep for 300 seconds......
    Sleep  180
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}/${image}
    Retry Wait Until Page Contains Element  ${not_scanned_icon}

    Log To Console  Sleep for 500 seconds......
    Sleep  500
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}/${image}
    Scan Result Should Display In List Row  ${sha256}
    View Repo Scan Details  Critical  High

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
    Create A New Endpoint    harbor    e${d}    https://cicd.harbor.vmwarecna.net    ${null}    ${null}    Y
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
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}/${image_a}
    Artifact Exist  ${tag_a}
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}/${image_b}
    Artifact Exist  ${tag_b}

    # Delete repository
    Go Into Project  ${project_name}
    Delete Repo  ${project_name}  ${image_a}
    Delete Repo  ${project_name}  ${image_b}

    # After replication schedule is set, project should contain 2 images.
    Log To Console  Sleep for 600 seconds......
    Sleep  600
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}/${image_a}
    Artifact Exist  ${tag_a}
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}/${image_b}
    Artifact Exist  ${tag_b}
