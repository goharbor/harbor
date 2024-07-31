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
Create An New Project And Go Into Project
    [Arguments]  ${projectname}  ${public}=false  ${count_quota}=${null}  ${storage_quota}=${null}  ${storage_quota_unit}=${null}  ${proxy_cache}=${false}  ${registry}=${null}
    Navigate To Projects
    FOR  ${n}  IN RANGE  1  8
        ${out}  Run Keyword And Ignore Error  Retry Button Click  xpath=${create_project_button_xpath}
        Log All  Return value is ${out[0]}
        Exit For Loop If  '${out[0]}'=='PASS'
    END
    Log To Console  Project Name: ${projectname}
    Retry Text Input  xpath=${project_name_xpath}  ${projectname}
    ${element_project_public}=  Set Variable  xpath=${project_public_xpath}
    Run Keyword If  '${public}' == 'true'  Run Keywords  Wait Until Element Is Visible And Enabled  ${element_project_public}  AND  Retry Element Click  ${element_project_public}
    Run Keyword If  '${count_quota}'!='${null}'  Input Count Quota  ${count_quota}
    Run Keyword If  '${storage_quota}'!='${null}'  Input Storage Quota  ${storage_quota}  ${storage_quota_unit}
    Run Keyword If  '${proxy_cache}' == '${true}'  Run Keywords  Retry Element Click  ${project_proxy_cache_switcher_xpath}  AND  Retry Element Click  ${project_registry_select_id}  AND  Retry Element Click  xpath=//select[@id='registry']//option[contains(.,'${registry}')]
    Retry Double Keywords When Error  Retry Element Click  ${create_project_OK_button_xpath}  Retry Wait Until Page Not Contains Element  ${create_project_OK_button_xpath}
    Go Into Project  ${projectname}  has_image=${false}

Create An New Project With New User
    [Arguments]  ${url}  ${username}  ${email}  ${realname}  ${newPassword}  ${comment}  ${projectname}  ${public}
    Create An New User  url=${url}  username=${username}  email=${email}  realname=${realname}  newPassword=${newPassword}  comment=${comment}
    Logout Harbor
    Sign In Harbor  ${url}  ${username}  ${newPassword}
    Create An New Project And Go Into Project  ${projectname}  ${public}

Artifact Exist
    [Arguments]  ${tag_name}
    Retry Wait Until Page Contains Element  //artifact-list-tab//clr-datagrid//clr-dg-row[contains(.,'sha256') and contains(.,'${tag_name}')]

Go To Project Log
    #Switch To Project Tab Overflow
    Retry Element Click  xpath=${project_log_xpath}

Switch To Member
    Retry Element Click  xpath=${project_member_xpath}

Switch To Replication
    Retry Element Click  xpath=${project_replication_xpath}

Switch To Project Configuration
    Retry Element Click  ${project_config_tabsheet}

Switch To Tag Retention
    #Switch To Project Tab Overflow
    Retry Element Click  xpath=${project_tag_strategy_xpath}

Switch To Tag Immutability
    #Switch To Project Tab Overflow
    Retry Double Keywords When Error  Retry Element Click  xpath=${project_tag_strategy_xpath}  Retry Wait Until Page Contains Element  ${project_tag_immutability_switch}
    Retry Double Keywords When Error  Retry Element Click  xpath=${project_tag_immutability_switch}  Retry Wait Until Page Contains  Immutability rules

Switch To Project Tab Overflow
    Retry Element Click  xpath=${project_tab_overflow_btn}

Navigate To Projects
    Retry Element Click  xpath=${projects_xpath}

Project Should Display
    [Arguments]  ${projectname}
    Retry Wait Element  xpath=//projects//list-project//clr-dg-cell/a[contains(.,'${projectname}')]

Project Should Not Display
    [Arguments]  ${projectname}
    Retry Wait Until Page Not Contains Element  xpath=//projects//list-project//clr-dg-cell/a[contains(.,'${projectname}')]

Search Private Projects
    Retry Element Click  xpath=//select
    Retry Element Click  xpath=//select/option[@value=1]

Make Project Private
    [Arguments]  ${projectname}
    Go Into Project  ${project name}
    Switch To Project Configuration
    Retry Checkbox Should Be Selected  ${project_config_public_checkbox}
    Retry Double Keywords When Error  Retry Element Click  ${project_config_public_checkbox_label}  Retry Checkbox Should Not Be Selected  ${project_config_public_checkbox}
    Retry Element Click  //button[contains(.,'SAVE')]
    Go Into Project  ${project name}
    Switch To Project Configuration
    Retry Checkbox Should Not Be Selected  ${project_config_public_checkbox}

Make Project Public
    [Arguments]  ${projectname}
    Go Into Project  ${project name}
    Switch To Project Configuration
    Retry Checkbox Should Not Be Selected  ${project_config_public_checkbox}
    Retry Double Keywords When Error  Retry Element Click  ${project_config_public_checkbox_label}  Retry Checkbox Should Be Selected  ${project_config_public_checkbox}
    Retry Element Click  //button[contains(.,'SAVE')]
    Go Into Project  ${project name}
    Switch To Project Configuration
    Retry Checkbox Should Be Selected  ${project_config_public_checkbox}

Repo Exist
    [Arguments]  ${pro_name}  ${repo_name}
    Retry Wait Until Page Contains Element  //clr-dg-row[contains(.,'${pro_name}/${repo_name}')]

Repo Not Exist
    [Arguments]  ${pro_name}  ${repo_name}
    Retry Wait Until Page Not Contains Element  //clr-dg-row[contains(.,'${pro_name}/${repo_name}')]

Filter Repo
    [Arguments]  ${pro_name}  ${repo_name}  ${exsit}=${true}
    ${filter_dist_input_visible}=  Run Keyword and Return Status  Element Should Not Be Visible  ${filter_dist_input}
    IF  ${filter_dist_input_visible}
        Retry Double Keywords When Error  Retry Element Click  ${filter_dist_btn}  Wait Until Element Is Visible And Enabled  ${filter_dist_input}
    END
    Retry Clear Element Text  ${filter_dist_input}
    Retry Text Input  ${filter_dist_input}  ${pro_name}/${repo_name}
    Run Keyword If  ${exsit}==${true}    Repo Exist  ${pro_name}  ${repo_name}
    ...  ELSE  Repo Not Exist  ${pro_name}  ${repo_name}

Delete Repo
    [Arguments]  ${pro_name}  ${repo_name}
    ${element_repo_checkbox}=  Set Variable  xpath=//clr-dg-row[contains(.,'${pro_name}/${repo_name}')]//div[contains(@class,'clr-checkbox-wrapper')]//label[contains(@class,'clr-control-label')]
    Filter Repo  ${pro_name}  ${repo_name}
    Retry Double Keywords When Error  Retry Element Click  ${element_repo_checkbox}  Wait Until Element Is Visible And Enabled  ${repo_delete_btn}
    Retry Double Keywords When Error  Retry Element Click  ${repo_delete_btn}  Wait Until Element Is Visible And Enabled  ${delete_confirm_btn}
    Retry Double Keywords When Error  Retry Element Click  ${delete_confirm_btn}  Retry Wait Until Page Not Contains Element  ${delete_confirm_btn}
    Retry Wait Until Page Not Contains Element  ${element_repo_checkbox}
    Filter Repo  ${pro_name}  ${repo_name}  exsit=${false}

Delete Repo on CardView
    [Arguments]  ${reponame}
    Retry Element Click  //hbr-gridview//span[contains(.,'${reponame}')]//clr-dropdown/button
    Retry Element Click  //hbr-gridview//span[contains(.,'${reponame}')]//clr-dropdown/clr-dropdown-menu/button[contains(.,'Delete')]
    Retry Element Click  ${repo_delete_on_card_view_btn}

Delete Project
    [Arguments]  ${projectname}
    Navigate To Projects
    Retry Element Click  xpath=//clr-dg-row[contains(.,'${projectname}')]//div[contains(@class,'clr-checkbox-wrapper')]//label[contains(@class,'clr-control-label')]
    Retry Element Click  ${project_action_xpath}
    Retry Element Click  xpath=//*[@id='delete-project']
    Retry Element Click  //clr-modal//button[contains(.,'DELETE')]

Project Should Not Be Deleted
    [Arguments]  ${projname}
    Delete Project  ${projname}
    Retry Wait Until Page Contains Element  //*[@id='contentAll']//div[contains(.,'${projname}')]/../div/clr-icon[@shape='error-standard']

Project Should Be Deleted
    [Arguments]  ${projname}
    Delete Project  ${projname}
    Retry Wait Until Page Contains Element  //*[@id='contentAll']//div[contains(.,'${projname}')]/../div/clr-icon[@shape='success-standard']

Advanced Search Should Display
    Retry Wait Until Page Contains Element  xpath=//audit-log//div[@class='flex-xs-middle']/button

# it's not a common keywords, only used into log case.
Do Log Advanced Search
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'artifact') and contains(.,'pull')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'artifact') and contains(.,'create')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'artifact') and contains(.,'delete')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'project') and contains(.,'create')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'repository') and contains(.,'delete')]
    Retry Element Click  xpath=//audit-log//div[@class='flex-xs-middle']/button
    Retry Element Click  xpath=//project-detail//audit-log//clr-dropdown/button
    #pull log
    Retry Element Click  xpath=//audit-log//clr-dropdown//a[contains(.,'Pull')]
    Retry Wait Until Page Not Contains Element  xpath=//clr-dg-row[contains(.,'pull')]
    #create log
    Retry Element Click  xpath=//audit-log//clr-dropdown/button
    Retry Element Click  xpath=//audit-log//clr-dropdown//a[contains(.,'Create')]
    Retry Wait Until Page Not Contains Element  xpath=//clr-dg-row[contains(.,'create')]
    #delete log
    Retry Element Click  xpath=//audit-log//clr-dropdown/button
    Retry Element Click  xpath=//audit-log//clr-dropdown//a[contains(.,'Delete')]
    Retry Wait Until Page Not Contains Element  xpath=//clr-dg-row[contains(.,'delete')]
    #others
    Retry Element Click  xpath=//audit-log//clr-dropdown/button
    Retry Element Click  xpath=//audit-log//clr-dropdown//a[contains(.,'Others')]
    Retry Element Click  xpath=//audit-log//hbr-filter//clr-icon
    Retry Text Input  xpath=//audit-log//hbr-filter//input  harbor-jobservice
    Retry Wait Until Page Not Contains Element   //audit-log//clr-dg-row[2]

Retry Click Repo Name
    [Arguments]  ${repo_name_element}
    FOR  ${n}  IN RANGE  1  2
        ${out}  Run Keyword And Ignore Error  Retry Double Keywords When Error  Retry Element Click  ${repo_name_element}   Retry Wait Element  ${tag_table_column_vulnerabilities}
        Exit For Loop If  '${out[0]}'=='PASS'
    END
    Should Be Equal As Strings  '${out[0]}'  'PASS'

    FOR  ${n}  IN RANGE  1  2
        ${out}  Run Keyword And Ignore Error  Retry Wait Until Page Not Contains Element  ${repo_list_spinner}
        Exit For Loop If  '${out[0]}'=='PASS'
    END
    Should Be Equal As Strings  '${out[0]}'  'PASS'

Go Into Project
    [Arguments]  ${project}  ${has_image}=${true}
    FOR  ${n}  IN RANGE  1  4
        ${out}  Run Keyword And Ignore Error  Retry Go Into Project  ${project}  ${has_image}
        Run Keyword If  '${out[0]}'=='PASS'  Exit For Loop
        Reload Page
        Sleep  2
    END
    Run Keyword If  '${out[0]}'=='FAIL'  Capture Page Screenshot
    Should Be Equal As Strings  '${out[0]}'  'PASS'

Retry Go Into Project
    [Arguments]  ${project}  ${has_image}
    Retry Text Input  ${search_input}  ${project}
    Wait Until Page Contains Element  //list-project-ro//a[contains(., '${project}')]
    Retry Link Click  //list-project-ro//a[contains(., '${project}')]
    Wait Until Page Contains Element  //project-detail//h1[contains(., '${project}')]
    Run Keyword If  ${has_image}==${false}  Wait Until Element Is Visible And Enabled  //clr-dg-placeholder[contains(.,\"We couldn\'t find any repositories!\")]
    ...  ELSE  Wait Until Element Is Visible And Enabled  //project-detail//hbr-repository-gridview//clr-dg-cell[contains(.,'${project}/')]

Go Into Repo
    [Arguments]  ${project_name}  ${repo_name}
    FOR  ${n}  IN RANGE  1  4
        ${out}  Run Keyword And Ignore Error  Retry Go Into Repo  ${project_name}  ${repo_name}
        Run Keyword If  '${out[0]}'=='PASS'  Exit For Loop
        Reload Page
        Sleep  2
    END
    Run Keyword If  '${out[0]}'=='FAIL'  Capture Page Screenshot
    Should Be Equal As Strings  '${out[0]}'  'PASS'

Retry Go Into Repo
    [Arguments]  ${project_name}  ${repo_name}
    Retry Text Input  ${search_input}  ${project_name}/${repo_name}
    Wait Until Page Contains Element  //list-repository-ro//a[contains(., '${project_name}/${repo_name}')]
    Retry Link Click  //list-repository-ro//a[contains(., '${project_name}/${repo_name}')]
    Wait Until Page Contains Element  //artifact-list-page//h2[contains(., '${repo_name}')]

Click Index Achieve
    [Arguments]  ${tag_name}
    Retry Element Click  //artifact-list-tab//clr-datagrid//clr-dg-row[contains(.,'sha256') and contains(.,'${tag_name}')]//clr-dg-cell[1]//clr-tooltip//a

Go Into Index And Contain Artifacts
    [Arguments]  ${tag_name}  ${total_artifact_count}=3  ${archive_count}=0  ${return_immediately}=${false}
    Run Keyword If  '${total_artifact_count}' == '${null}'  Return From Keyword   PASS
    Should Not Be Empty  ${tag_name}
    Retry Double Keywords When Error  Click Index Achieve  ${tag_name}  Page Should Contain Element  ${tag_table_column_os_arch}
    FOR  ${n}  IN RANGE  1  10
        ${out1}  Run Keyword And Ignore Error  Page Should Contain Element  ${artifact_rows}  limit=${total_artifact_count}
        ${out2}  Run Keyword And Ignore Error  Page Should Contain Element  ${archive_rows}  limit=${archive_count}
        Exit For Loop If  '${out1[0]}'=='PASS' and '${out2[0]}'=='PASS'
        Sleep  3
    END
    ${result}=  Set Variable If  '${out1[0]}'=='FAIL' or '${out2[0]}'=='FAIL'  FAIL  PASS
    Return From Keyword If  '${return_immediately}' == '${true}'  ${result}
    Should Be Equal As Strings  '${out1[0]}'  'PASS'
    Should Be Equal As Strings  '${out2[0]}'  'PASS'

Switch To CardView
    Retry Element Click  xpath=//hbr-repository-gridview//span[@class='card-btn']/clr-icon

Expand Repo
    [Arguments]  ${projectname}
    Retry Element Click  //repository//clr-dg-row[contains(.,'${projectname}')]//button/clr-icon

Edit Repo Info
    Retry Element Click  //*[@id='repo-info']
    Retry Wait Until Page Contains Element  //*[@id='info']/form/div[2]
    # Cancel input
    Retry Element Click  xpath=//*[@id='info-edit-button']/button
    Input Text  xpath=//*[@id='info-edit-textarea']  test_description_info
    Retry Element Click  xpath=//*[@id='edit-cancel']
    Retry Element Click  xpath=//clr-modal//button[contains(.,'CONFIRM')]
    Retry Wait Until Page Contains Element  //*[@id='no-editing']
    # Confirm input
    Retry Element Click  xpath=//*[@id='info-edit-button']/button
    Input Text  xpath=//*[@id='info-edit-textarea']  test_description_info
    Retry Element Click  xpath=//*[@id='edit-save']
    Retry Wait Until Page Contains  test_description_info

Switch To Project Label
    Retry Element Click  xpath=//project-detail//a[contains(.,'Labels')]

Switch To Project Repo
    Retry Element Click  xpath=//project-detail//a[contains(.,'Repositories')]

Switch To Project Scanner
    Retry Element Click  xpath=//project-detail//a[contains(.,'Scanner')]

Add Labels To Tag
    [Arguments]  ${tagName}  ${labelName}
    Retry Element Click  xpath=//clr-dg-row[contains(.,'${tagName}')]//label[contains(@class,'clr-control-label')]
    Retry Element Click  xpath=//clr-dg-action-bar//clr-dropdown//span
    Retry Element Click  xpath=//clr-dropdown-menu//clr-dropdown//button[contains(.,'Add Labels')]
    Retry Element Click  xpath=//clr-dropdown//div//label[contains(.,'${labelName}')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row//label[contains(.,'${labelName}')]

Filter Labels In Tags
    [Arguments]  ${labelName1}  ${labelName2}
    Retry Element Click  xpath=//*[@id='search-btn']
    Retry Element Click  xpath=//*[@id='type-select']
    Retry Element Click  xpath=//*[@id='type-select']/option[@value='labels']
    Retry Wait Until Page Contains Element  xpath=//*[@id='filterArea']//div//button[contains(.,'${labelName1}')]
    Retry Element Click  xpath=//*[@id='filterArea']//div//button[contains(.,'${labelName1}')]
    Retry Element Click  xpath=//app-artifact-filter//clr-icon[contains(@shape,'search')]
    Retry Wait Until Page Contains Element  xpath=//clr-datagrid//label[contains(.,'${labelName1}')]

    Retry Element Click  xpath=//*[@id='search-btn']
    Retry Element Click  xpath=//*[@id='filterArea']//div//button[contains(.,'${labelName2}')]
    Retry Element Click  xpath=//app-artifact-filter//clr-icon[contains(@shape,'search')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'${labelName2}')]
    Retry Wait Until Page Not Contains Element  xpath=//clr-dg-row[contains(.,'${labelName1}')]

Get Statics
    [Arguments]  ${locator}
    Reload Page
    Wait Until Element Is Visible And Enabled  ${locator}
    ${privaterepo}=  Get Text  ${locator}
    [Return]  ${privaterepo}

Retry Get Statics
    [Arguments]  ${locator}
    @{param}  Create List  ${locator}
    ${ret}=  Retry Keyword N Times When Error  5  Get Statics  @{param}
    [Return]  ${ret}

Get Statics Private Repo
    ${privaterepo}=  Retry Get Statics  ${project_statistics_private_repository_icon}
    [Return]  ${privaterepo}

Get Statics Private Project
    ${privateproj}=  Retry Get Statics  //projects/div/div/div[1]/div/statistics-panel/div/div[1]/div/div[1]/div[2]
    [Return]  ${privateproj}

Get Statics Public Repo
    ${publicrepo}=  Retry Get Statics  //projects/div/div/div[1]/div/statistics-panel/div/div[1]/div/div[2]/div[2]
    [Return]  ${publicrepo}

Get Statics Public Project
    ${publicproj}=  Retry Get Statics  //projects/div/div/div[1]/div/statistics-panel/div/div[1]/div/div[2]/div[2]
    [Return]  ${publicproj}

Get Statics Total Repo
    ${totalrepo}=  Retry Get Statics  //projects/div/div/div[1]/div/statistics-panel/div/div[2]/div/div[3]/div[2]
    [Return]  ${totalrepo}

Get Statics Total Project
    ${totalproj}=  Retry Get Statics  //projects/div/div/div[1]/div/statistics-panel/div/div[1]/div/div[3]/div[2]
    [Return]  ${totalproj}

Input Count Quota
    [Arguments]  ${text}
    ${element_xpath}=  Set Variable  ${project_add_count_quota_input_text_id}
    Retry Clear Element Text  ${element_xpath}
    Retry Text Input  ${element_xpath}  ${text}

Input Storage Quota
    [Arguments]  ${text}  ${unit}=${null}
    ${element_xpath}=  Set Variable  ${project_add_storage_quota_input_text_id}
    Retry Clear Element Text  ${element_xpath}
    Retry Text Input  ${element_xpath}  ${text}
    Run Keyword If  '${unit}'!='${null}'  Select Storage Quota unit  ${unit}

Select Storage Quota unit
    [Arguments]  ${unit}
    Select From List By Value  ${project_add_storage_quota_unit_id}  ${unit}

Back Project Home
    [Arguments]  ${project_name}
    Retry Link Click  //a[contains(.,'${project_name}')]

Should Be Signed
    [Arguments]  ${tag}
    Retry Wait Element Visible  //clr-dg-row[contains(.,'${tag}')]//clr-icon[contains(@class,'signed')]

Should Not Be Signed
    [Arguments]  ${tag}
    Retry Wait Element Visible  //clr-dg-row[contains(.,'${tag}')]//clr-icon[contains(@class,'color-red')]

Should Be Signed By Cosign
    [Arguments]  ${tag}=${null}  ${digest}=${null}
    IF  '${tag}' != '${null}'
        Retry Wait Element Visible  //clr-dg-row[./clr-expandable-animation/div/div/div/clr-dg-cell/div/clr-tooltip/div/div/span[contains(.,'${tag}')] and .//clr-dg-row[.//img[@title='signature.cosign']]]
    ELSE
        Retry Wait Element Visible  //clr-dg-row[./clr-expandable-animation/div/div/div/clr-dg-cell/div/a[contains(.,'${digest}')] and .//clr-dg-row[.//img[@title='signature.cosign']]]
    END

Should Be Signed By Notation
    [Arguments]  ${tag}=${null}  ${digest}=${null}
    IF  '${tag}' != '${null}'
        Retry Wait Element Visible  //clr-dg-row[./clr-expandable-animation/div/div/div/clr-dg-cell/div/clr-tooltip/div/div/span[contains(.,'${tag}')] and .//clr-dg-row[.//img[@title='signature.notation']]]
    ELSE
        Retry Wait Element Visible  //clr-dg-row[./clr-expandable-animation/div/div/div/clr-dg-cell/div/a[contains(.,'${digest}')] and .//clr-dg-row[.//img[@title='signature.notation']]]
    END

Delete Accessory
    [Arguments]  ${tag}
    Retry Button Click  //clr-dg-row[contains(.,'${tag}')]//button[contains(@class,'datagrid-expandable-caret-button')]
    Retry Button Click  //clr-dg-row[contains(.,'${tag}')]//button[contains(@class,'datagrid-action-toggle')]
    Retry Button Click  //button[contains(.,'Delete')]
    Retry Button Click  //div[contains(@class,'modal-content')]//button[contains(@class,'btn-danger')]

Should be Accessory deleted
    [Arguments]  ${tag}
    Retry Wait Until Page Not Contains Element  //clr-dg-row[contains(.,'${tag}')]//button[contains(@class,'datagrid-expandable-caret-button')]

Export CVEs
    [Arguments]  ${project}  ${repositories}  ${tags}  ${labels}  ${cve_ids}
    Filter Project  ${project}
    Retry Element Click  //clr-dg-row[contains(.,'${project}')]//div[contains(@class,'clr-checkbox-wrapper')]//label[contains(@class,'clr-control-label')]
    Retry Element Click  ${project_action_xpath}
    Retry Button Click  ${export_cve_btn}
    Retry Text Input  ${export_cve_filter_repo_input}  ${repositories}
    Retry Text Input  ${export_cve_filter_tag_input}  ${tags}
    Select Filter Label For CVE Export  @{labels}
    Retry Text Input  ${export_cve_filter_cveid_input}  ${cve_ids}
    Retry Double Keywords When Error  Retry Button Click  ${export_btn}  Retry Wait Until Page Contains  Trigger exporting CVEs successfully!

Should Not Be Export CVEs
     Retry Element Click  ${project_action_xpath}
     Retry Wait Element Should Be Disabled  ${export_cve_btn}
     Retry Element Click  ${project_action_xpath}

Download Latest CVE CSV File
    Retry Element Click  ${event_log_xpath}
    ${csv_file_name}=  Get Text  ${latest_cve_csv_file_name_xpath}
    ${csv_file_path}=  Set Variable  ${download_directory}/${csv_file_name}.csv
    Retry Double Keywords When Error  Retry Element Click  ${latest_download_cve_csv_file_xpath}  Retry Wait Until Page Does Not Contains  ${csv_file_name}
    Retry File Should Exist  ${csv_file_path}
    [Return]  ${csv_file_path}

Select Project Scanner
    [Arguments]  ${scanner_name}  ${scanner_count}=${null}
    Retry Element Click  //*[@id='edit-scanner']
    Run Keyword If  '${scanner_count}'!='${null}'   Retry Wait Element Count  //clr-dg-row  ${scanner_count}
    Retry Element Click  //clr-dg-row[.//clr-dg-cell[text()='${scanner_name}']]//label[contains(@class,'clr-control-label')]
    Retry Element Click  //*[@id='save-scanner']
    Retry Wait Element Visible  //span[@id='scanner-name' and text()='${scanner_name}']
