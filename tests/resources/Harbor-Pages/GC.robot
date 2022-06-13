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
Switch To Garbage Collection
    Retry Double Keywords When Error  Retry Element Click  xpath=${gc_page_xpath}  Retry Wait Until Page Contains Element  ${gc_now_button}

GC Now
    [Arguments]  ${untag}=${false}
    Switch To Garbage Collection
    Run Keyword If  '${untag}' == '${true}'  Retry Element Click  xpath=${checkbox_delete_untagged_artifacts}
    Retry Double Keywords When Error  Retry Element Click  ${gc_now_button}  Retry Wait Until Page Contains  Running

Retry GC Should Be Successful
    [Arguments]  ${history_id}  ${expected_msg}
    Retry Keyword N Times When Error  15  GC Should Be Successful  ${history_id}  ${expected_msg}

GC Should Be Successful
    [Arguments]  ${history_id}  ${expected_msg}
    ${rc}  ${output}=  Run And Return Rc And Output  curl -u ${HARBOR_ADMIN}:${HARBOR_PASSWORD} -i --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/v2.0/system/gc/${history_id}/log"
    Log All  ${output}
    Should Be Equal As Integers  ${rc}  0
    Run Keyword If  '${expected_msg}' != '${null}'  Should Contain  ${output}  ${expected_msg}
    Should Contain  ${output}  success to run gc in job.

Get GC Logs
    ${cmd}=  Set Variable  curl -u ${HARBOR_ADMIN}:${HARBOR_PASSWORD} -s --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/v2.0/system/gc"
    Log All  cmd:${cmd}
    ${rc}  ${output}=  Run And Return Rc And Output  ${cmd}
    Log All  ${output}
    [Return]  ${output}

Set GC Schedule
    [Arguments]  ${type}  ${value}=${null}
    Switch To Garbage Collection
    Retry Double Keywords When Error  Retry Element Click  ${gc_schedule_edit_btn}  Retry Wait Until Page Not Contains Element  ${gc_schedule_edit_btn}
    Retry Element Click  ${GC_schedule_select}
    Run Keyword If  '${type}'=='custom'  Run Keywords  Retry Element Click  ${vulnerability_dropdown_list_item_custom}  AND  Retry Text Input  ${targetCron_id}  ${value}
    ...  ELSE  Retry Element Click  ${vulnerability_dropdown_list_item_none}
    Retry Double Keywords When Error  Retry Element Click  ${GC_schedule_save_btn}  Retry Wait Until Page Not Contains Element  ${gc_schedule_save_btn}
    Capture Page Screenshot