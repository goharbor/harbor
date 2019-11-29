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
Create An New Project
    [Arguments]  ${projectname}  ${public}=false  ${count_quota}=${null}  ${storage_quota}=${null}  ${storage_quota_unit}=${null}
    Navigate To Projects
    Retry Button Click  xpath=${create_project_button_xpath}
    Log To Console  Project Name: ${projectname}
    Capture Page Screenshot
    Retry Text Input  xpath=${project_name_xpath}  ${projectname}
    ${element_project_public}=  Set Variable  xpath=${project_public_xpath}
    Run Keyword If  '${public}' == 'true'  Run Keywords  Wait Until Element Is Visible And Enabled  ${element_project_public}  AND  Click Element  ${element_project_public}
    Run Keyword If  '${count_quota}'!='${null}'  Input Count Quota  ${count_quota}
    Run Keyword If  '${storage_quota}'!='${null}'  Input Storage Quota  ${storage_quota}  ${storage_quota_unit}
    Capture Page Screenshot
    Retry Double Keywords When Error  Retry Element Click  ${create_project_OK_button_xpath}  Retry Wait Until Page Not Contains Element  ${create_project_OK_button_xpath}
    Capture Page Screenshot
    Go Into Project  ${projectname}  has_image=${false}

Create An New Project With New User
    [Arguments]  ${url}  ${username}  ${email}  ${realname}  ${newPassword}  ${comment}  ${projectname}  ${public}
    Create An New User  url=${url}  username=${username}  email=${email}  realname=${realname}  newPassword=${newPassword}  comment=${comment}
    Logout Harbor
    Sign In Harbor  ${url}  ${username}  ${newPassword}
    Create An New Project  ${projectname}  ${public}
    Sleep  1

#It's the log of project.
Go To Project Log
    Retry Element Click  xpath=${project_log_xpath}
    Sleep  2

Switch To Member
    Sleep  3
    Retry Element Click  xpath=${project_member_xpath}
    Sleep  1

Switch To Log
    Retry Element Click  xpath=${log_xpath}
    Sleep  1

Switch To Replication
    Retry Element Click  xpath=${project_replication_xpath}
    Sleep  1

Switch To Project Configuration
    Retry Element Click  ${project_config_tabsheet}
    Sleep  1

Switch To Tag Retention
    Retry Element Click  xpath=${project_tag_retention_xpath}
    Sleep  1

Navigate To Projects
    Retry Element Click  xpath=${projects_xpath}
    Sleep  2

Project Should Display
    [Arguments]  ${projectname}
    Retry Wait Element  xpath=//project//list-project//clr-dg-cell/a[contains(.,'${projectname}')]

Project Should Not Display
    [Arguments]  ${projectname}
    Retry Wait Until Page Not Contains Element  xpath=//project//list-project//clr-dg-cell/a[contains(.,'${projectname}')]

Search Private Projects
    Retry Element Click  xpath=//select
    Retry Element Click  xpath=//select/option[@value=1]
    Sleep  1
    Capture Page Screenshot  SearchPrivateProjects.png

Make Project Private
    [Arguments]  ${projectname}
    Go Into Project  ${project name}
    Switch To Project Configuration
    Retry Checkbox Should Be Selected  ${project_config_public_checkbox}
    Retry Double Keywords When Error  Retry Element Click  ${project_config_public_checkbox_label}  Retry Checkbox Should Not Be Selected  ${project_config_public_checkbox}
    Retry Element Click  //button[contains(.,'SAVE')]
    Retry Wait Until Page Contains  Configuration has been successfully saved

Make Project Public
    [Arguments]  ${projectname}
    Go Into Project  ${project name}
    Switch To Project Configuration
    Retry Checkbox Should Not Be Selected  ${project_config_public_checkbox}
    Retry Double Keywords When Error  Retry Element Click  ${project_config_public_checkbox_label}  Retry Checkbox Should Be Selected  ${project_config_public_checkbox}
    Retry Element Click  //button[contains(.,'SAVE')]
    Retry Wait Until Page Contains  Configuration has been successfully saved

Delete Repo
    [Arguments]  ${projectname}
    ${element_repo_checkbox}=  Set Variable  xpath=//clr-dg-row[contains(.,'${projectname}')]//clr-checkbox-wrapper//label
    Retry Double Keywords When Error  Retry Element Click  ${element_repo_checkbox}  Wait Until Element Is Visible And Enabled  ${repo_delete_btn}
    Retry Double Keywords When Error  Retry Element Click  ${repo_delete_btn}  Wait Until Element Is Visible And Enabled  ${delete_confirm_btn}
    Retry Double Keywords When Error  Retry Element Click  ${delete_confirm_btn}  Retry Wait Until Page Not Contains Element  ${delete_confirm_btn}
    Retry Wait Until Page Not Contains Element  ${element_repo_checkbox}

Delete Repo on CardView
    [Arguments]  ${reponame}
    Retry Element Click  //hbr-gridview//span[contains(.,'${reponame}')]//clr-dropdown/button
    Retry Element Click  //hbr-gridview//span[contains(.,'${reponame}')]//clr-dropdown/clr-dropdown-menu/button[contains(.,'Delete')]
    Retry Element Click  ${repo_delete_on_card_view_btn}
    Sleep  2

Delete Project
    [Arguments]  ${projectname}
    Navigate To Projects
    Retry Element Click  xpath=//clr-dg-row[contains(.,'${projectname}')]//clr-checkbox-wrapper//label
    Retry Element Click  xpath=//*[@id='delete-project']
    Retry Element Click  //clr-modal//button[contains(.,'DELETE')]
    Sleep  1

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
    Capture Page Screenshot  LogAdvancedSearch.png
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'pull')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'push')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'create')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'delete')]
    Retry Element Click  xpath=//audit-log//div[@class='flex-xs-middle']/button
    Retry Element Click  xpath=//project-detail//audit-log//clr-dropdown/button
    #pull log
    Retry Element Click  xpath=//audit-log//clr-dropdown//a[contains(.,'Pull')]
    Retry Wait Until Page Not Contains Element  xpath=//clr-dg-row[contains(.,'pull')]
    #push log
    Retry Element Click  xpath=//audit-log//clr-dropdown/button
    Retry Element Click  xpath=//audit-log//clr-dropdown//a[contains(.,'Push')]
    Retry Wait Until Page Not Contains Element  xpath=//clr-dg-row[contains(.,'push')]
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
    Retry Text Input  xpath=//audit-log//hbr-filter//input  harbor
    Sleep  1
    Capture Page Screenshot  LogAdvancedSearch2.png
    ${rc} =  Get Matching Xpath Count  //audit-log//clr-dg-row
    Should Be Equal As Integers  ${rc}  0

Go Into Repo
    [Arguments]  ${repoName}
    Sleep  2
    Retry Wait Until Page Not Contains Element  ${repo_list_spinner}
    ${repo_name_element}=  Set Variable  xpath=//clr-dg-cell[contains(.,'${repoName}')]/a
    Retry Element Click  ${repo_search_icon}
    :For  ${n}  IN RANGE  1  10
    \    Retry Clear Element Text  ${repo_search_input}
    \    Retry Text Input  ${repo_search_input}  ${repoName}
    \    ${out}  Run Keyword And Ignore Error  Retry Wait Until Page Contains Element  ${repo_name_element}
    \    Exit For Loop If  '${out[0]}'=='PASS'
    \    Sleep  2
    Capture Page Screenshot
    Retry Double Keywords When Error  Retry Element Click  ${repo_name_element}  Retry Wait Until Page Not Contains Element  ${repo_name_element}
    Capture Page Screenshot
    Retry Wait Element  ${tag_table_column_pull_command}
    Retry Wait Element  ${tag_images_btn}
    Capture Page Screenshot

Switch To CardView
    Retry Element Click  xpath=//hbr-repository-gridview//span[@class='card-btn']/clr-icon
    Sleep  5

Expand Repo
    [Arguments]  ${projectname}
    Retry Element Click  //repository//clr-dg-row[contains(.,'${projectname}')]//button/clr-icon
    Sleep  1

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
    Capture Page Screenshot

Switch To Project Label
    Retry Element Click  xpath=//project-detail//a[contains(.,'Labels')]
    Sleep  1

Switch To Project Repo
    Retry Element Click  xpath=//project-detail//a[contains(.,'Repositories')]
    Sleep  1

Add Labels To Tag
    [Arguments]  ${tagName}  ${labelName}
    Retry Element Click  xpath=//clr-dg-row[contains(.,'${tagName}')]//label
    Capture Page Screenshot  add_${labelName}.png
    Retry Element Click  xpath=//clr-dg-action-bar//clr-dropdown//button
    Retry Element Click  xpath=//clr-dropdown//div//label[contains(.,'${labelName}')]
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row//label[contains(.,'${labelName}')]

Filter Labels In Tags
    [Arguments]  ${labelName1}  ${labelName2}
    Retry Element Click  xpath=//*[@id='filterArea']//hbr-filter/span/clr-icon
    Retry Wait Until Page Contains Element  xpath=//*[@id='filterArea']//div//button[contains(.,'${labelName1}')]
    Retry Element Click  xpath=//*[@id='filterArea']//div//button[contains(.,'${labelName1}')]
    Retry Element Click  xpath=//*[@id='filterArea']//hbr-filter/span/clr-icon
    Retry Wait Until Page Contains Element  xpath=//clr-datagrid//label[contains(.,'${labelName1}')]

    Retry Element Click  xpath=//*[@id='filterArea']//hbr-filter/span/clr-icon
    Retry Element Click  xpath=//*[@id='filterArea']//div//button[contains(.,'${labelName2}')]
    Retry Element Click  xpath=//*[@id='filterArea']//hbr-filter/span/clr-icon
    Sleep  2
    Capture Page Screenshot  filter_${labelName2}.png
    Retry Wait Until Page Contains Element  xpath=//clr-dg-row[contains(.,'${labelName2}')]
    Retry Wait Until Page Not Contains Element  xpath=//clr-dg-row[contains(.,'${labelName1}')]

Get Statics Private Repo
    ${privaterepo}=  Get Text  //project/div/div/div[1]/div/statistics-panel/div/div[2]/div[1]/div[2]/div[2]/statistics/div/span[1]
    Convert To Integer  ${privaterepo}
    [Return]  ${privaterepo}

Get Statics Private Project
    ${privateproj}=  Get Text  //project/div/div/div[1]/div/statistics-panel/div/div[2]/div[1]/div[2]/div[1]/statistics/div/span[1]
    Convert To Integer  ${privateproj}
    [Return]  ${privateproj}

Get Statics Public Repo
    ${publicrepo}=  Get Text  //project/div/div/div[1]/div/statistics-panel/div/div[2]/div[1]/div[3]/div[2]/statistics/div/span[1]
    Convert To Integer  ${publicrepo}
    [Return]  ${publicrepo}

Get Statics Public Project
    ${publicproj}=  Get Text  //project/div/div/div[1]/div/statistics-panel/div/div[2]/div[1]/div[3]/div[1]/statistics/div/span[1]
    Convert To Integer  ${publicproj}
    [Return]  ${publicproj}

Get Statics Total Repo
    ${totalrepo}=  Get Text  //project/div/div/div[1]/div/statistics-panel/div/div[2]/div[1]/div[4]/div[2]/statistics/div/span[1]
     Convert To Integer  ${totalrepo}
    [Return]  ${totalrepo}

Get Statics Total Project
    ${totalproj}=  Get Text  //project/div/div/div[1]/div/statistics-panel/div/div[2]/div[1]/div[4]/div[1]/statistics/div/span[1]
    Convert To Integer  ${totalproj}
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

