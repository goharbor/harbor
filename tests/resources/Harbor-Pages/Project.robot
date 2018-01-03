# Copyright 2016-2017 VMware, Inc. All Rights Reserved.
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
    Wait Until Page Contains  ${projectname}
    Wait Until Page Contains  Project Admin

Create An New Project With New User
    [Arguments]  ${url}  ${username}  ${email}  ${realname}  ${newPassword}  ${comment}  ${projectname}  ${public}
    Create An New User  url=${url}  username=${username}  email=${email}  realname=${realname}  newPassword=${newPassword}  comment=${comment}
    Logout Harbor
    Sign In Harbor  ${url}  ${username}  ${newPassword}
    Create An New Project  ${projectname}  ${public}
    Sleep  1	

#It's the log of project.
Go To Project Log
    Click Element  xpath=//project-detail//ul/li[3]
    Sleep  2

Switch To Member
    Click Element  xpath=//project-detail//li[2]
    Sleep  1

Switch To Log
    Click Element  xpath=${log_xpath}
    Sleep  1

Switch To Replication
    Click Element  xpath=${replication_xpath}
    Sleep  1

Back To projects
    Click Element  xpath=${projects_xpath}
    Sleep  1

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
    #Click element  xpath=//project//list-project//clr-dg-row-master[contains(.,'${projectname}')]//clr-dg-action-overflow
    #Click element  xpath=//project//list-project//clr-dg-action-overflow//button[contains(.,"Make Private")]
    Checkbox Should Be Selected  xpath=//input[@name='public']
    Click Element  //clr-checkbox[@name='public']//label
    Click Element  //button[contains(.,'SAVE')]

Make Project Public
    [Arguments]  ${projectname}
    Go Into Project  ${project name}    
    Sleep  1
    Click Element  xpath=//project-detail//a[contains(.,'Configuration')]
    #Click element  xpath=//project//list-project//clr-dg-row-master[contains(.,'${projectname}')]//clr-dg-action-overflow
    #Click element  xpath=//project//list-project//clr-dg-action-overflow//button[contains(.,"Make Public")]
    Checkbox Should Not Be Selected  xpath=//input[@name='public']
    Click Element  //clr-checkbox[@name='public']//label
    Click Element  //button[contains(.,'SAVE')]

Delete Repo
    [Arguments]  ${projectname}
    #Click Element  xpath=//project-detail//clr-dg-row-master[contains(.,"${projectname}")]//clr-dg-action-overflow
    Click Element  xpath=//clr-dg-row[contains(.,"${projectname}")]//clr-checkbox//label
    Sleep  1
    Click Element  xpath=//button[contains(.,"Delete")]
    Sleep  1
    Click Element  xpath=//clr-modal//button[2]
    Sleep  1
    Click Element  xpath=//button[contains(.,"CLOSE")]

Delete Project
    [Arguments]  ${projectname}
    Sleep  1
    #Click Element  //list-project//clr-dg-row-master[contains(.,'${projname}')]//clr-dg-action-overflow
    #Click Element  //list-project//clr-dg-row-master[contains(.,'${projname}')]//clr-dg-action-overflow//button[contains(.,'Delete')]
    #click delete button to confirm
    Click Element  xpath=//clr-dg-row[contains(.,"${projectname}")]//clr-checkbox//label
    Sleep  1
    Click Element  xpath=//button[contains(.,"Delete")]
    Sleep  1
    Click Element  xpath=//clr-modal//button[2]
    Sleep  1
    Click Element  xpath=//button[contains(.,"CLOSE")]

Project Should Not Be Deleted
    [Arguments]  ${projname}
    Delete Project  ${projname}
    Sleep  1
    Page Should Contain  ${projname}

Project Should Be Deleted
    [Arguments]  ${projname}
    Delete Project  ${projname}
    Sleep  2
    Page Should Not Contain  ${projname}

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
    Click Element  xpath=//*[@id="search_input"]
    Sleep  2
    Input Text  xpath=//*[@id="search_input"]  ${repoName}
    Sleep  8
    Wait Until Page Contains  ${repoName}
    Click Element  xpath=//*[@id="results"]/list-repository-ro//clr-dg-cell[contains(.,${repoName})]/a
    Sleep  2
    Capture Page Screenshot  gointo_${repoName}.png

Expand Repo
    [Arguments]  ${projectname}
    Click Element  //repository//clr-dg-row[contains(.,'${projectname}')]//button/clr-icon
    Sleep  1

Scan Repo
    [Arguments]  ${tagname}
    #Click Element  //hbr-tag//clr-dg-row-master[contains(.,'${tagname}')]//clr-dg-action-overflow
    #Click Element  //hbr-tag//clr-dg-row-master[contains(.,'${tagname}')]//clr-dg-action-overflow//button[contains(.,'Scan')]
    #select one tag
    Click Element  //clr-dg-row[contains(.,"${tagname}")]//label
    Click Element  //button[contains(.,'Scan')]
    Sleep  15

Edit Repo Info
    Click Element  //*[@id="repo-info"]
    Sleep  1
    Page Should Contain Element  //*[@id="info"]/form/div[2]/h3
    # Cancel input
    Click Element  xpath=//*[@id="info-edit-button"]/button
    Input Text  xpath=//*[@id="info"]/form/div[2]/textarea  test_description_info
    Click Element  xpath=//*[@id="info"]/form/div[3]/button[2]
    Sleep  1
    Click Element  xpath=//*[@id="info"]/form/confirmation-dialog/clr-modal/div/div[1]/div/div[1]/div/div[3]/button[2]
    Sleep  1
    Page Should Contain Element  //*[@id="info"]/form/div[2]/h3
    # Confirm input
    Click Element  xpath=//*[@id="info-edit-button"]/button
    Input Text  xpath=//*[@id="info"]/form/div[2]/textarea  test_description_info
    Click Element  xpath=//*[@id="info"]/form/div[3]/button[1]
    Sleep  1
    Page Should Contain  test_description_info
    Capture Page Screenshot  RepoInfo.png

Summary Chart Should Display
    [Arguments]  ${tagname}
    Page Should Contain Element  //clr-dg-row[contains(.,'${tagname}')]//hbr-vulnerability-bar//hbr-vulnerability-summary-chart
