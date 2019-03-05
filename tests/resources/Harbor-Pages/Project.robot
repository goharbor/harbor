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
${HARBOR_VERSION}  v1.1.1

*** Keywords ***
Create An New Project
    [Arguments]  ${projectname}  ${public}=false
    Navigate To Projects
    ${element_create_project_button}=  Set Variable  xpath=${create_project_button_xpath}
    Wait Until Element Is Visible And Enabled  ${element_create_project_button}
    Click Button  ${element_create_project_button}
    Log To Console  Project Name: ${projectname}
    ${elemen_project_name}=  Set Variable  xpath=${project_name_xpath}
    Wait Until Element Is Visible And Enabled  ${elemen_project_name}
    Input Text  ${elemen_project_name}  ${projectname}
    ${element_project_public}=  Set Variable  xpath=${project_public_xpath}
    Run Keyword If  '${public}' == 'true'  Run Keywords  Wait Until Element Is Visible And Enabled  ${element_project_public}  AND  Click Element  ${element_project_public}
    ${element_create_project_OK_button_xpath}=  Set Variable  ${create_project_OK_button_xpath}
    Wait Until Element Is Visible And Enabled  ${element_create_project_OK_button_xpath}
    Click Element  ${element_create_project_OK_button_xpath}
    Wait Until Page Does Not Contain Element  ${create_project_CANCEL_button_xpath}
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
    Click Element  xpath=${project_log_xpath}
    Sleep  2

Switch To Member
    Sleep  3
    Click Element  xpath=${project_member_xpath}
    Sleep  1

Switch To Log
    Wait Until Element Is Enabled  xpath=${log_xpath}
    Wait Until Element Is Visible  xpath=${log_xpath}
    Click Element  xpath=${log_xpath}
    Sleep  1

Switch To Replication
    Click Element  xpath=${project_replication_xpath}
    Sleep  1

Navigate To Projects
    ${element}=  Set Variable  xpath=${projects_xpath}
    Wait Until Element Is Visible And Enabled  ${element}
    Click Element  ${element}
    Sleep  2

Project Should Display
    [Arguments]  ${projectname}
    ${element}=  Set Variable  xpath=//project//list-project//clr-dg-cell/a[contains(.,'${projectname}')]
    Wait Until Element Is Visible And Enabled  ${element}

Project Should Not Display
    [Arguments]  ${projectname}
    Page Should Not Contain Element  xpath=//project//list-project//clr-dg-cell/a[contains(.,'${projectname}')]

Search Private Projects
    Click element  xpath=//select
    Click element  xpath=//select/option[@value=1]
    Sleep  1
    Capture Page Screenshot  SearchPrivateProjects.png

Make Project Private
    [Arguments]  ${projectname}
    Go Into Project  ${project name}
    Sleep  2
    Click Element  xpath=//project-detail//a[contains(.,'Configuration')]
    Sleep  1
    Checkbox Should Be Selected  xpath=//input[@name='public']
    Click Element  //div[@id="clr-wrapper-public"]//label[1]
    Wait Until Element Is Enabled  //button[contains(.,'SAVE')]
    Click Element  //button[contains(.,'SAVE')]
    Wait Until Page Contains  Configuration has been successfully saved

Make Project Public
    [Arguments]  ${projectname}
    Go Into Project  ${project name}
    Retry Element Click  xpath=//project-detail//a[contains(.,'Configuration')]
    Checkbox Should Not Be Selected  xpath=//input[@name='public']
    Retry Element Click  //div[@id="clr-wrapper-public"]//label[1]
    Retry Element Click  //button[contains(.,'SAVE')]
    Wait Until Page Contains  Configuration has been successfully saved

Delete Repo
    [Arguments]  ${projectname}
    ${element_repo_checkbox}=  Set Variable  xpath=//clr-dg-row[contains(.,'${projectname}')]//clr-checkbox-wrapper//label
    Retry Double Keywords When Error  Retry Element Click  ${element_repo_checkbox}  Wait Until Element Is Visible And Enabled  ${repo_delete_btn}
    Retry Double Keywords When Error  Retry Element Click  ${repo_delete_btn}  Wait Until Element Is Visible And Enabled  ${delete_confirm_btn}
    Retry Double Keywords When Error  Retry Element Click  ${delete_confirm_btn}  Retry Wait Until Page Not Contains Element  ${delete_confirm_btn}

Delete Repo on CardView
    [Arguments]  ${reponame}
    Retry Element Click  //hbr-gridview//span[contains(.,'${reponame}')]//clr-dropdown/button
    Retry Element Click  //hbr-gridview//span[contains(.,'${reponame}')]//clr-dropdown/clr-dropdown-menu/button[contains(.,'Delete')]
    Retry Element Click  ${repo_delete_on_card_view_btn}
    Sleep  2

Delete Project
    [Arguments]  ${projectname}
    Navigate To Projects
    Sleep  1
    Click Element  xpath=//clr-dg-row[contains(.,'${projectname}')]//clr-checkbox-wrapper//label
    Sleep  1
    Click Element  xpath=//button[contains(.,'Delete')]
    Sleep  2
    Click Element  //clr-modal//button[contains(.,'DELETE')]
    Sleep  1

Project Should Not Be Deleted
    [Arguments]  ${projname}
    Delete Project  ${projname}
    Sleep  1
    Page Should Contain Element  //clr-tab-content//div[contains(.,'${projname}')]/../div/clr-icon[@shape='error-standard']

Project Should Be Deleted
    [Arguments]  ${projname}
    Delete Project  ${projname}
    Sleep  2
    Page Should Contain Element  //clr-tab-content//div[contains(.,'${projname}')]/../div/clr-icon[@shape='success-standard']

Advanced Search Should Display
    Page Should Contain Element  xpath=//audit-log//div[@class='flex-xs-middle']/button

# it's not a common keywords, only used into log case.
Do Log Advanced Search
    Capture Page Screenshot  LogAdvancedSearch.png
    Sleep  1
    Page Should Contain Element  xpath=//clr-dg-row[contains(.,'pull')]
    Page Should Contain Element  xpath=//clr-dg-row[contains(.,'push')]
    Page Should Contain Element  xpath=//clr-dg-row[contains(.,'create')]
    Page Should Contain Element  xpath=//clr-dg-row[contains(.,'delete')]
    Sleep  1
    Click Element  xpath=//audit-log//div[@class='flex-xs-middle']/button
    Sleep  1
    Click Element  xpath=//project-detail//audit-log//clr-dropdown/button
    Sleep  1
    #pull log
    Sleep  1
    Click Element  xpath=//audit-log//clr-dropdown//a[contains(.,'Pull')]
    Sleep  1
    Page Should Not Contain Element  xpath=//clr-dg-row[contains(.,'pull')]
    #push log
    Click Element  xpath=//audit-log//clr-dropdown/button
    Sleep  1
    Click Element  xpath=//audit-log//clr-dropdown//a[contains(.,'Push')]
    Sleep  1
    Page Should Not Contain Element  xpath=//clr-dg-row[contains(.,'push')]
    #create log
    Click Element  xpath=//audit-log//clr-dropdown/button
    Sleep  1
    Click Element  xpath=//audit-log//clr-dropdown//a[contains(.,'Create')]
    Sleep  1
    Page Should Not Contain Element  xpath=//clr-dg-row[contains(.,'create')]
    #delete log
    Click Element  xpath=//audit-log//clr-dropdown/button
    Sleep  1
    Click Element  xpath=//audit-log//clr-dropdown//a[contains(.,'Delete')]
    Sleep  1
    Page Should Not Contain Element  xpath=//clr-dg-row[contains(.,'delete')]
    #others
    Click Element  xpath=//audit-log//clr-dropdown/button
    Click Element  xpath=//audit-log//clr-dropdown//a[contains(.,'Others')]
    Sleep  1
    Click Element  xpath=//audit-log//hbr-filter//clr-icon
    Input Text  xpath=//audit-log//hbr-filter//input  harbor
    Sleep  1
    Capture Page Screenshot  LogAdvancedSearch2.png
    ${rc} =  Get Matching Xpath Count  //audit-log//clr-dg-row
    Should Be Equal As Integers  ${rc}  0

Go Into Repo
    [Arguments]  ${repoName}
    Sleep  2
    Click Element  xpath=//hbr-filter//clr-icon
    Sleep  2
    Input Text  xpath=//hbr-filter//input  ${repoName}
    Sleep  3
    Wait Until Page Contains  ${repoName}
    Click Element  xpath=//clr-dg-cell[contains(.,${repoName})]/a
    Sleep  2
    Capture Page Screenshot  gointo_${repoName}.png

Switch To CardView
    Sleep  2
    Click Element  xpath=//hbr-repository-gridview//span[@class='card-btn']/clr-icon
    Sleep  5

Expand Repo
    [Arguments]  ${projectname}
    Click Element  //repository//clr-dg-row[contains(.,'${projectname}')]//button/clr-icon
    Sleep  1

Edit Repo Info
    Click Element  //*[@id='repo-info']
    Sleep  1
    Page Should Contain Element  //*[@id='info']/form/div[2]
    # Cancel input
    Click Element  xpath=//*[@id='info-edit-button']/button
    Input Text  xpath=//*[@id='info']/form/div[2]/textarea  test_description_info
    Click Element  xpath=//*[@id='info']/form/div[3]/button[2]
    Sleep  1
    Click Element  xpath=//*[@id='info']/form/confirmation-dialog/clr-modal/div/div[1]/div[1]/div/div[3]/button[2]
    Sleep  1
    Page Should Contain Element  //*[@id='info']/form/div[2]
    # Confirm input
    Click Element  xpath=//*[@id='info-edit-button']/button
    Input Text  xpath=//*[@id='info']/form/div[2]/textarea  test_description_info
    Click Element  xpath=//*[@id='info']/form/div[3]/button[1]
    Sleep  1
    Page Should Contain  test_description_info
    Capture Page Screenshot  RepoInfo.png

Switch To Project Label
    Click Element  xpath=//project-detail//a[contains(.,'Labels')]
    Sleep  1

Switch To Project Repo
    Click Element  xpath=//project-detail//a[contains(.,'Repositories')]
    Sleep  1

Add Labels To Tag
    [Arguments]  ${tagName}  ${labelName}
    Click Element  xpath=//clr-dg-row[contains(.,'${tagName}')]//label
    Capture Page Screenshot  add_${labelName}.png
    Sleep  1
    Click Element  xpath=//clr-dg-action-bar//clr-dropdown//button
    Sleep  1
    Click Element  xpath=//clr-dropdown//div//label[contains(.,'${labelName}')]
    Sleep  3
    Page Should Contain Element  xpath=//clr-dg-row//label[contains(.,'${labelName}')]

Filter Labels In Tags
    [Arguments]  ${labelName1}  ${labelName2}
    Sleep  2
    Click Element  xpath=//*[@id='filterArea']//hbr-filter/span/clr-icon
    Sleep  2
    Page Should Contain Element  xpath=//*[@id='filterArea']//div//button[contains(.,'${labelName1}')]
    Click Element  xpath=//*[@id='filterArea']//div//button[contains(.,'${labelName1}')]
    Sleep  2
    Click Element  xpath=//*[@id='filterArea']//hbr-filter/span/clr-icon
    Page Should Contain Element  xpath=//clr-datagrid//label[contains(.,'${labelName1}')]

    Click Element  xpath=//*[@id='filterArea']//hbr-filter/span/clr-icon
    Sleep  2
    Click Element  xpath=//*[@id='filterArea']//div//button[contains(.,'${labelName2}')]
    Sleep  2
    Click Element  xpath=//*[@id='filterArea']//hbr-filter/span/clr-icon
    Sleep  2
    Capture Page Screenshot  filter_${labelName2}.png
    Page Should Contain Element  xpath=//clr-dg-row[contains(.,'${labelName2}')]
    Page Should Not Contain Element  xpath=//clr-dg-row[contains(.,'${labelName1}')]

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

