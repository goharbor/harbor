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
    Sleep  1
    Click Button  css=${create_project_button_css}
    Sleep  1
    Log To Console  Project Name: ${projectname}
    Input Text  xpath=${project_name_xpath}  ${projectname}
    Sleep  3
    Run Keyword If  '${public}' == 'true'  Click Element  xpath=${project_public_xpath}
    Click Element  xpath=//button[contains(.,'OK')]
    Sleep  4
    ${rc}  ${output}=  Run And Return Rc And Output  curl -u ${HARBOR_ADMIN}:${HARBOR_PASSWORD} -k -X GET --header 'Accept: application/json' ${HARBOR_URL}/api/projects?name=${projectname}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  ${projectname}

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
    Click Element  xpath=${project_member_xpath}
    Sleep  1

Switch To Log
    Click Element  xpath=${log_xpath}
    Sleep  1

Switch To Replication
    Click Element  xpath=${project_replication_xpath}
    Sleep  1

Back To projects
    Click Element  xpath=${projects_xpath}
    Sleep  2

Project Should Display
    [Arguments]  ${projectname}
    Page Should Contain Element  xpath=//project//list-project//clr-dg-cell/a[contains(.,'${projectname}')]

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
    Sleep  1
    Click Element  xpath=//project-detail//a[contains(.,'Configuration')]
    Sleep  1
    Checkbox Should Be Selected  xpath=//input[@name='public']
    Click Element  //clr-checkbox[@name='public']//label
    Wait Until Element Is Enabled  //button[contains(.,'SAVE')]
    Click Element  //button[contains(.,'SAVE')]
    Wait Until Page Contains  Configuration has been successfully saved

Make Project Public
    [Arguments]  ${projectname}
    Go Into Project  ${project name}    
    Sleep  1
    Click Element  xpath=//project-detail//a[contains(.,'Configuration')]
    Checkbox Should Not Be Selected  xpath=//input[@name='public']
    Click Element  //clr-checkbox[@name='public']//label
    Wait Until Element Is Enabled  //button[contains(.,'SAVE')]
    Click Element  //button[contains(.,'SAVE')]
    Wait Until Page Contains  Configuration has been successfully saved

Delete Repo
    [Arguments]  ${projectname}
    Click Element  xpath=//clr-dg-row[contains(.,"${projectname}")]//clr-checkbox//label
    Wait Until Element Is Enabled  //button[contains(.,"Delete")]
    Click Element  xpath=//button[contains(.,"Delete")]
    Wait Until Element Is Visible  //clr-modal//button[2]
    Click Element  xpath=//clr-modal//button[2]

Delete Repo on CardView
    [Arguments]  ${reponame}
    Click Element  //hbr-gridview//span[contains(.,'${reponame}')]//clr-dropdown/button
    Wait Until Element Is Visible  //hbr-gridview//span[contains(.,'${reponame}')]//clr-dropdown/clr-dropdown-menu/button[contains(.,'Delete')]
    Click Element  //hbr-gridview//span[contains(.,'${reponame}')]//clr-dropdown/clr-dropdown-menu/button[contains(.,'Delete')]
    Wait Until Element Is Visible  //clr-modal//button[contains(.,'DELETE')]
    Click Element  //clr-modal//button[contains(.,'DELETE')]

Delete Project
    [Arguments]  ${projectname}
    Sleep  1
    Click Element  xpath=//clr-dg-row[contains(.,"${projectname}")]//clr-checkbox//label
    Sleep  1
    Click Element  xpath=//button[contains(.,"Delete")]
    Sleep  2
    Click Element  //clr-modal//button[contains(.,'DELETE')]
    Sleep  1

Project Should Not Be Deleted
    [Arguments]  ${projname}
    Delete Project  ${projname}
    Sleep  1
    Page Should Contain Element  //clr-tab-content//div[contains(.,'${projname}')]/../div/clr-icon[@shape="error-standard"]

Project Should Be Deleted
    [Arguments]  ${projname}
    Delete Project  ${projname}
    Sleep  2
    Page Should Contain Element  //clr-tab-content//div[contains(.,'${projname}')]/../div/clr-icon[@shape="success-standard"]

Advanced Search Should Display
    Page Should Contain Element  xpath=//audit-log//div[@class="flex-xs-middle"]/button

# it's not a common keywords, only used into log case.	
Do Log Advanced Search
    Capture Page Screenshot  LogAdvancedSearch.png
    Sleep  1
    Page Should Contain Element  xpath=//clr-dg-row[contains(.,"pull")]
    Page Should Contain Element  xpath=//clr-dg-row[contains(.,"push")]
    Page Should Contain Element  xpath=//clr-dg-row[contains(.,"create")]
    Page Should Contain Element  xpath=//clr-dg-row[contains(.,"delete")]
    Sleep  1
    Click Element  xpath=//audit-log//div[@class="flex-xs-middle"]/button
    Sleep  1
    Click Element  xpath=//project-detail//audit-log//clr-dropdown/button
    Sleep  1
    #pull log
    Sleep  1
    Click Element  xpath=//audit-log//clr-dropdown//a[contains(.,"Pull")]
    Sleep  1
    Page Should Not Contain Element  xpath=//clr-dg-row[contains(.,"pull")]
    #push log
    Click Element  xpath=//audit-log//clr-dropdown/button
    Sleep  1
    Click Element  xpath=//audit-log//clr-dropdown//a[contains(.,"Push")]
    Sleep  1
    Page Should Not Contain Element  xpath=//clr-dg-row[contains(.,"push")]
    #create log
    Click Element  xpath=//audit-log//clr-dropdown/button
    Sleep  1
    Click Element  xpath=//audit-log//clr-dropdown//a[contains(.,"Create")]
    Sleep  1
    Page Should Not Contain Element  xpath=//clr-dg-row[contains(.,"create")]
    #delete log
    Click Element  xpath=//audit-log//clr-dropdown/button
    Sleep  1
    Click Element  xpath=//audit-log//clr-dropdown//a[contains(.,"Delete")]
    Sleep  1
    Page Should Not Contain Element  xpath=//clr-dg-row[contains(.,"delete")]
    #others
    Click Element  xpath=//audit-log//clr-dropdown/button
    Click Element  xpath=//audit-log//clr-dropdown//a[contains(.,"Others")]
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
    Click Element  //*[@id="repo-info"]
    Sleep  1
    Page Should Contain Element  //*[@id="info"]/form/div[2]
    # Cancel input
    Click Element  xpath=//*[@id="info-edit-button"]/button
    Input Text  xpath=//*[@id="info"]/form/div[2]/textarea  test_description_info
    Click Element  xpath=//*[@id="info"]/form/div[3]/button[2]
    Sleep  1
    Click Element  xpath=//*[@id="info"]/form/confirmation-dialog/clr-modal/div/div[1]/div/div[1]/div/div[3]/button[2]
    Sleep  1
    Page Should Contain Element  //*[@id="info"]/form/div[2]
    # Confirm input
    Click Element  xpath=//*[@id="info-edit-button"]/button
    Input Text  xpath=//*[@id="info"]/form/div[2]/textarea  test_description_info
    Click Element  xpath=//*[@id="info"]/form/div[3]/button[1]
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
    Click Element  xpath=//clr-dg-row[contains(.,"${tagName}")]//label
    Capture Page Screenshot  add_${labelName}.png
    Sleep  1
    Click Element  xpath=//clr-dg-action-bar//clr-dropdown//button
    Sleep  1
    Click Element  xpath=//clr-dropdown//div//label[contains(.,"${labelName}")]
    Sleep  3
    Page Should Contain Element  xpath=//clr-dg-row//label[contains(.,"${labelName}")]

Filter Labels In Tags
    [Arguments]  ${labelName1}  ${labelName2}
    Sleep  2
    Click Element  xpath=//*[@id="filterArea"]//hbr-filter/span/clr-icon
    Sleep  2
    Page Should Contain Element  xpath=//*[@id="filterArea"]//div//button[contains(.,"${labelName1}")]
    Click Element  xpath=//*[@id="filterArea"]//div//button[contains(.,"${labelName1}")]
    Sleep  2
    Click Element  xpath=//*[@id="filterArea"]//hbr-filter/span/clr-icon
    Page Should Contain Element  xpath=//clr-datagrid//label[contains(.,"${labelName1}")]

    Click Element  xpath=//*[@id="filterArea"]//hbr-filter/span/clr-icon
    Sleep  2
    Click Element  xpath=//*[@id="filterArea"]//div//button[contains(.,"${labelName2}")]
    Sleep  2
    Click Element  xpath=//*[@id="filterArea"]//hbr-filter/span/clr-icon
    Sleep  2
    Capture Page Screenshot  filter_${labelName2}.png
    Page Should Contain Element  xpath=//clr-dg-row[contains(.,"${labelName2}")]
    Page Should Not Contain Element  xpath=//clr-dg-row[contains(.,"${labelName1}")]

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

