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
Switch To Security Hub
    Retry Element Click  xpath=//clr-main-container//clr-vertical-nav//a[contains(.,'Interrogation')]
    Retry Element Click  xpath=//app-interrogation-services//a[contains(.,'Security Hub')]
    Retry Wait Element  ${security_hub_search_btn}

Get Vulnerability System Summary From API
    ${cmd}=  Set Variable  curl -u ${HARBOR_ADMIN}:${HARBOR_PASSWORD} -s --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/v2.0/security/summary?with_dangerous_cve=true&with_dangerous_artifact=true"
    ${rc}  ${output}=  Run And Return Rc And Output  ${cmd}
    ${output_json}  Evaluate  json.loads('''${output}''')  json
    [Return]  ${output_json}

Check The Total Vulnerabilities
    [Arguments]  ${summary}
    Retry Wait Element  (//div[@class='card'][1]//div[contains(@class, 'clr-col-9')])[1][text()=' ${summary["critical_cnt"]} ']
    Retry Wait Element  (//div[@class='card'][1]//div[contains(@class, 'clr-col-9')])[2][text()=' ${summary["high_cnt"]} ']
    Retry Wait Element  (//div[@class='card'][1]//div[contains(@class, 'clr-col-9')])[3][text()=' ${summary["medium_cnt"]} ']
    Retry Wait Element  (//div[@class='card'][1]//div[contains(@class, 'clr-col-9')])[4][text()=' ${summary["low_cnt"]} ']
    Retry Wait Element  (//div[@class='card'][1]//div[contains(@class, 'clr-col-9')])[5][text()=' 0 ']
    Retry Wait Element  (//div[@class='card'][1]//div[contains(@class, 'clr-col-9')])[6][text()=' 0 ']

Check The Top 5 Most Dangerous Artifacts
    [Arguments]  ${dangerous_artifacts}
    Retry Wait Element Count  ${top5_most_dangerous_artifacts_xpath}  5
    FOR  ${index}  ${dangerous_artifact}  IN ENUMERATE  @{dangerous_artifacts}  start=1
        ${repository_name}=  Set Variable  ${dangerous_artifact["repository_name"]}
        ${short_digest}=  Set Variable  ${dangerous_artifact["digest"]}[0:15]
        ${row_num}=  Set Variable  [${index}]
        ${text}=  Set Variable  [..//a[@title='${repository_name}'] and ..//span[text()='${short_digest}']]
        Wait Until Element Is Visible And Enabled  ${top5_most_dangerous_artifacts_xpath}${row_num}${text}
    END

Check The Top 5 Most Dangerous CVEs
    [Arguments]  ${dangerous_cves}
    Retry Wait Element Count  ${top5_most_dangerous_cves_xpath}  5
    FOR  ${index}  ${dangerous_cve}  IN ENUMERATE  @{dangerous_cves}  start=1
        ${dangerous_cve_id}=  Set Variable  ${dangerous_cve["cve_id"]}
        ${cvss_score_v3}=  Set Variable  ${dangerous_cve["cvss_score_v3"]}
        ${dangerous_cve_package}=  Set Variable  ${dangerous_cve["package"]}\@${dangerous_cve["version"]}
        ${severity}=  Set Variable  ${dangerous_cve["severity"]}
        ${row_num}=  Set Variable  [${index}]
        ${text}=  Set Variable  [..//a[@title='${dangerous_cve_id}'] and ..//span[text()='${severity}'] and ..//div[text()=' ${cvss_score_v3} '] and ..//span[text()=' ${dangerous_cve_package} ']]
        Wait Until Element Is Visible And Enabled  ${top5_most_dangerous_cves_xpath}${row_num}${text}
        IF  ${index} < 5
            ${next_cvss_score_v3}=  Get From Dictionary  ${dangerous_cves}[${index}]  cvss_score_v3
            ${comparison_result}=  Evaluate  ${cvss_score_v3} >= ${next_cvss_score_v3}
            Should Be True  ${comparison_result}
        END
    END

Check The Search By One Condition
    [Arguments]  ${project_name}  ${repository_name}  ${digest}  ${cve_id}  ${package}  ${tag}  ${cvss_score_v3_from}  ${cvss_score_v3_to}  ${summary}
    # Check the search by project name
    Select From List By Value  ${vulnerabilities_filter_select}  project_id
    Retry Text Input  ${vulnerabilities_filter_input}  ${project_name}
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  //div[@class='datagrid']//clr-dg-cell[2][starts-with(@title, '${project_name}')]  10
    # Check the search by repository name
    Select From List By Value  ${vulnerabilities_filter_select}  repository_name
    Retry Text Input  ${vulnerabilities_filter_input}  ${repository_name}
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  //div[@class='datagrid']//clr-dg-cell[2][@title='${repository_name}']  10
    # Check the search by artifact digest
    Select From List By Value  ${vulnerabilities_filter_select}  digest
    Retry Text Input  ${vulnerabilities_filter_input}  ${digest}
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  //div[@class='datagrid']//clr-dg-cell[3][@title='${digest}']  10
    ${short_digest}=  Set Variable  ${digest}[0:15]
    Retry Wait Element Count  //div[@class='datagrid']//clr-dg-cell[3]//a[text()='${short_digest}']  10
    # Check the search by CVE ID
    Select From List By Value  ${vulnerabilities_filter_select}  cve_id
    Retry Text Input  ${vulnerabilities_filter_input}  ${cve_id}
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  //div[@class='datagrid']//clr-dg-cell[1]//a[text()='${cve_id}']  10
    # Check the search by package
    Select From List By Value  ${vulnerabilities_filter_select}  package
    Retry Text Input  ${vulnerabilities_filter_input}  ${package}
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  //div[@class='datagrid']//clr-dg-cell[7][@title='${package}']  10
    # Check the search by tag
    Select From List By Value  ${vulnerabilities_filter_select}  tag
    Retry Text Input  ${vulnerabilities_filter_input}  ${tag}
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  //div[@class='datagrid']//clr-dg-cell[4][text()='${tag}']  10
    # Check the search by CVSS3
    Select From List By Value  ${vulnerabilities_filter_select}  cvss_score_v3
    ${cvss3_from_input}=  Format String  {}{}  ${vulnerabilities_filter_input}  [1]
    ${cvss3_to_input}=  Format String  {}{}  ${vulnerabilities_filter_input}  [2]
    Retry Text Input  ${cvss3_from_input}  ${cvss_score_v3_from}
    Retry Text Input  ${cvss3_to_input}  ${cvss_score_v3_to}
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  //div[@class='datagrid']//clr-dg-cell[5][text()>=${cvss_score_v3_from} and text()<=${cvss_score_v3_to}]  10
    # Check the search by severity
    # Critical
    Select From List By Value  ${vulnerabilities_filter_select}  severity
    Select From List By Value  //form//div[2]//select  Critical
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  //div[@class='datagrid']//clr-dg-cell[6]//span[text()='Critical']  10
    Retry Wait Element  //clr-dg-footer//span[text()='${summary["critical_cnt"]} CVEs']
    # High
    Select From List By Value  //form//div[2]//select  High
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  //div[@class='datagrid']//clr-dg-cell[6]//span[text()='High']  10
    Retry Wait Element  //clr-dg-footer//span[text()='${summary["high_cnt"]} CVEs']
    # Medium
    Select From List By Value  //form//div[2]//select  Medium
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  //div[@class='datagrid']//clr-dg-cell[6]//span[text()='Medium']  10
    Retry Wait Element  //clr-dg-footer//span[text()='${summary["medium_cnt"]} CVEs']
    # Low
    Select From List By Value  //form//div[2]//select  Low
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  //div[@class='datagrid']//clr-dg-cell[6]//span[text()='Low']  10
    Retry Wait Element  //clr-dg-footer//span[text()='${summary["low_cnt"]} CVEs']
    # n/a
    Select From List By Value  //form//div[2]//select  Unknown
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  ${vulnerabilities_datagrid_row}  0
    Retry Wait Element  //clr-dg-footer//span[text()='0 CVEs']
    # None
    Select From List By Value  //form//div[2]//select  None
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  ${vulnerabilities_datagrid_row}  0
    Retry Wait Element  //clr-dg-footer//span[text()='0 CVEs']

Check The Search By All Condition
    [Arguments]  ${project_name}  ${repository_name}  ${digest}  ${cve_id}  ${package}  ${tag}  ${cvss_score_v3_from}  ${cvss_score_v3_to}  ${severity}
    # project name
    Select From List By Value  ${vulnerabilities_filter_select}  project_id
    Retry Text Input  ${vulnerabilities_filter_input}  ${project_name}
    Retry Wait Element  ${remove_search_criteria_icon_disabled}
    # repository name
    Retry Element Click  ${add_search_criteria_icon}
    ${repository_name_select}=  Format String  {}{}  ${vulnerabilities_filter_select}  [2]
    ${repository_name_input}=  Format String  {}{}  ${vulnerabilities_filter_input}  [2]
    Select From List By Value  ${repository_name_select}  repository_name
    Retry Text Input  ${repository_name_input}  ${repository_name}
    # artifact digest
    Retry Element Click  ${add_search_criteria_icon}
    ${digest_name_select}=  Format String  {}{}  ${vulnerabilities_filter_select}  [3]
    ${digest_name_input}=  Format String  {}{}  ${vulnerabilities_filter_input}  [3]
    Select From List By Value  ${digest_name_select}  digest
    Retry Text Input  ${digest_name_input}  ${digest}
    # CVE ID
    Retry Element Click  ${add_search_criteria_icon}
    ${cve_id_select}=  Format String  {}{}  ${vulnerabilities_filter_select}  [4]
    ${cve_id_input}=  Format String  {}{}  ${vulnerabilities_filter_input}  [4]
    Select From List By Value  ${cve_id_select}  cve_id
    Retry Text Input  ${cve_id_input}  ${cve_id}
    # package
    Retry Element Click  ${add_search_criteria_icon}
    ${package_select}=  Format String  {}{}  ${vulnerabilities_filter_select}  [5]
    ${package_input}=  Format String  {}{}  ${vulnerabilities_filter_input}  [5]
    Select From List By Value  ${package_select}  package
    Retry Text Input  ${package_input}  ${package}
    # tag
    Retry Element Click  ${add_search_criteria_icon}
    ${tag_select}=  Format String  {}{}  ${vulnerabilities_filter_select}  [6]
    ${tag_input}=  Format String  {}{}  ${vulnerabilities_filter_input}  [6]
    Select From List By Value  ${tag_select}  tag
    Retry Text Input  ${tag_input}  ${tag}
    # CVSS3
    Retry Element Click  ${add_search_criteria_icon}
    ${cvss3_select}=  Format String  {}{}  ${vulnerabilities_filter_select}  [7]
    ${cvss3_from_input}=  Format String  {}{}  ${vulnerabilities_filter_input}  [7]
    ${cvss3_to_input}=  Format String  {}{}  ${vulnerabilities_filter_input}  [8]
    Select From List By Value  ${cvss3_select}  cvss_score_v3
    Retry Text Input  ${cvss3_from_input}  ${cvss_score_v3_from}
    Retry Text Input  ${cvss3_to_input}  ${cvss_score_v3_to}
    # severity
    Retry Element Click  ${add_search_criteria_icon}
    Retry Wait Element  ${add_search_criteria_icon_disabled}
    Retry Wait Element  ${remove_search_criteria_icon}
    ${severity_select}=  Format String  {}{}  ${vulnerabilities_filter_select}  [8]
    ${severity_input}=  Format String  {}{}  (//form[contains(@class,'clr-form')]//select)  [9]
    Select From List By Value  ${severity_select}  severity
    Select From List By Value  ${severity_input}  ${severity}
    # search
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  ${vulnerabilities_datagrid_row}  1
    ${target_row_xpath}=  Set Variable  //div[@class='datagrid'][..//clr-dg-cell[2][@title='${repository_name}'] and ..//clr-dg-cell[3][@title='${digest}'] and ..//clr-dg-cell[1]//a[text()='${cve_id}'] and ..//clr-dg-cell[7][@title='${package}'] and ..//clr-dg-cell[4][text()='${tag}'] and ..//clr-dg-cell[5][text()>=${cvss_score_v3_from} and text()<=${cvss_score_v3_to}] and ..//clr-dg-cell[6]//span[text()='${severity}']]
    Log  ${target_row_xpath}
    Retry Wait Element  ${target_row_xpath}
    FOR  ${index}  IN RANGE  7
        Retry Element Click  ${remove_search_criteria_icon}
    END
    Retry Wait Element  ${remove_search_criteria_icon_disabled}
    Retry Wait Element  ${add_search_criteria_icon}

Check The Vulnerabilities Jump
    [Arguments]  ${project_name}  ${repository_name}  ${cve_id}  ${cve_description}
    Retry Wait Until Page Does Not Contains  ${cve_description}
    Retry Double Keywords When Error  Retry Button Click  //clr-dg-row//button  Retry Wait Until Page Contains  ${cve_description}
    Retry Double Keywords When Error  Retry Button Click  //clr-dg-row//button  Retry Wait Until Page Does Not Contains  ${cve_description}
    # Vulnerabilities datagrid CVE jump
    Retry Double Keywords When Error  Click Link New Tab And Switch  (//clr-dg-row//clr-dg-cell[1])[1]//a  Retry Wait Element  //h1[contains(.,'${cve_id}')]
    Switch Window  locator=MAIN
    # Vulnerabilities datagrid repository jump
    Retry Link Click  (//clr-dg-row//clr-dg-cell[2])[1]//a
    Retry Wait Element  //h2[text()=' ${repository_name} ']
    Retry Wait Element  //a[text()='${project_name}']
    Switch To Security Hub
    # Vulnerabilities datagrid digest jump
    Retry Wait Element  (//clr-dg-row//clr-dg-cell[3])[1]//a
    ${short_digest}=  Get Text  (//clr-dg-row//clr-dg-cell[3])[1]//a
    Retry Link Click  (//clr-dg-row//clr-dg-cell[3])[1]//a
    Retry Wait Element  //h2//span[text()='${short_digest}']
    Switch To Security Hub
    # Top 5 Most Dangerous Artifacts jump
    ${short_digest}=  Set Variable  sha256:415bfdcf
    Retry Element Click  //div[@class='card'][2]//span[text()='${short_digest}']
    Retry Wait Element  //h2//span[text()='${short_digest}']
    Switch To Security Hub
    # Top 5 Most Dangerous Artifacts jump
    ${short_digest}=  Set Variable  sha256:7bf979f2
    Retry Element Click  //div[@class='card'][2]//span[text()='${short_digest}']
    Retry Wait Element  //h2//span[text()='${short_digest}']

Check The Quick Search
    # Search for the most dangerous artifact
    ${repository_name_xpath}=  Set Variable  (//div[@class='card'][2]//span)[1]
    ${digest_xpath}=  Set Variable  (//div[@class='card'][2]//span)[2]
    Retry Wait Element  ${repository_name_xpath}
    Retry Wait Element  ${digest_xpath}
    ${repository_name}=  Get Text  ${repository_name_xpath}
    ${digest}=  Get Text  ${digest_xpath}
    Retry Element Click  ${repository_name_xpath}
    Retry Wait Element Count  ${vulnerabilities_filter_select}  2
    ${repository_name_select}=  Format String  {}{}  ${vulnerabilities_filter_select}  [1]
    ${repository_name_input}=  Format String  {}{}  ${vulnerabilities_filter_input}  [1]
    ${digest_select}=  Format String  {}{}  ${vulnerabilities_filter_select}  [2]
    ${digest_input}=  Format String  {}{}  ${vulnerabilities_filter_input}  [2]
    ${repository_name_selected}=  Get Selected List Value    ${repository_name_select}
    ${digest_selected}=  Get Selected List Value    ${digest_select}
    Should Be Equal As Strings  ${repository_name_selected}  repository_name
    Should Be Equal As Strings  ${digest_selected}  digest
    ${repository_name_input_value}=  Get Value  ${repository_name_input}
    ${digest_input_value}=  Get Value  ${digest_input}
    Should Be Equal As Strings  ${repository_name_input_value}  ${repository_name}
    Should Start With  ${digest_input_value}  ${digest}
    ${row_count}=  Get Element Count  ${vulnerabilities_datagrid_row}
    Retry Wait Element Count  //clr-datagrid//clr-dg-row[..//clr-dg-cell[2][@title='${repository_name}'] and ..//clr-dg-cell[3][starts-with(@title,'${digest}')]]  ${row_count}
    # Search for the most dangerous CVEs
    ${cve_xpath}=  Set Variable  (//div[@class='card'][3]//span)[1]
    ${cve}=  Get Text  ${cve_xpath}
    Retry Element Click  ${cve_xpath}
    Retry Wait Element Count  ${vulnerabilities_filter_select}  1
    ${cve_select}=  Format String  {}{}  ${vulnerabilities_filter_select}  [1]
    ${cve_input}=  Format String  {}{}  ${vulnerabilities_filter_input}  [1]
    ${cve_selected}=  Get Selected List Value    ${cve_select}
    Should Be Equal As Strings  ${cve_selected}  cve_id
    ${cve_input_value}=  Get Value  ${cve_input}
    Should Be Equal As Strings  ${cve_input_value}  ${cve}
    ${row_count}=  Get Element Count  ${vulnerabilities_datagrid_row}
    Retry Wait Element Count  //div[@class='datagrid']//clr-dg-cell[1]//a[text()='${cve}']  ${row_count}

Select Filter Label For CVE Export
    [Arguments]    @{labels}
    Retry Element Click  ${vulnerabilities_filter_label_xpath}
    FOR  ${label}  IN  @{labels}
        Log  ${label}
        Retry Element Click  //hbr-label-piece//span[contains(text(), '${label}')]
    END
    Retry Element Click  ${vulnerabilities_filter_label_xpath}
