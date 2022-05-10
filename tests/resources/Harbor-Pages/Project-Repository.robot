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

*** Keywords ***
View Repo Scan Details
    [Arguments]  @{vulnerabilities_level}
    Retry Element Click  xpath=${first_repo_xpath}
    FOR  ${item}  IN  @{vulnerabilities_level}
        Retry Wait Until Page Contains Element  //hbr-artifact-vulnerabilities//clr-dg-row[contains(.,'${item}')]
    END
    Retry Element Click  xpath=${build_history_btn}
    Retry Wait Until Page Contains Element  xpath=${build_history_data}

View Scan Error Log
    Retry Wait Until Page Contains  View Log
    Retry Element Click  xpath=${view_log_xpath}

Scan Artifact
    [Arguments]  ${project}  ${repo}  ${label_xpath}=//clr-dg-row//label[1]
    Go Into Project  ${project}
    Go Into Repo  ${project}/${repo}
    Retry Element Click  ${label_xpath}
    Retry Element Click  ${scan_artifact_btn}

Stop Scan Artifact
    Retry Element Click  ${stop_scan_artifact_btn}

Check Scan Artifact Job Status Is Stopped
    Wait Until Element Is Visible  ${stopped_label}
    ${job_status}=  Get Text  ${stopped_label}
    Should Be Equal As Strings  '${job_status}'  'Scan stopped'

Refresh Repositories
    Retry Element Click  ${refresh_repositories_xpath}