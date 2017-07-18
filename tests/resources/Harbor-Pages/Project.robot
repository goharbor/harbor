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
		[Arguments]  ${projectname}
		Sleep  1
		Click Button  css=${create_project_button_css}
		Sleep  1
		Log To Console  Project Name: ${projectname}
		Input Text  xpath=${project_name_xpath}  ${projectname}
		Sleep  3
		Capture Page Screenshot
		Click Element  css=${project_save_css}
		Sleep  4
		Wait Until Page Contains  ${projectname}
		Wait Until Page Contains  Project Admin
		Capture Page Screenshot

Create An New Public Project
		[Arguments]  ${projectname}
		Sleep  1
		Click Button  css=${create_project_button_css}
		Sleep  1
		Log To Console  Project Name: ${projectname}
		Input Text  xpath=${project_name_xpath}  ${projectname}
		Sleep  3
		Click Element  xpath=${project_public_xpath}
		Click Element  css=${project_save_css}
		Sleep  4
		Wait Until Page Contains  ${projectname}
		Wait Until Page Contains  Project Admin
		Capture Page Screenshot

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
		Sleep  1
		Click element  xpath=//project//list-project//clr-dg-row-master[contains(.,'${projectname}')]//clr-dg-action-overflow
		Click element  xpath=//project//list-project//clr-dg-action-overflow//button[contains(.,"Make Private")]

Make Project Public
		[Arguments]  ${projectname}
		Sleep  1
		Click element  xpath=//project//list-project//clr-dg-row-master[contains(.,'${projectname}')]//clr-dg-action-overflow
		Click element  xpath=//project//list-project//clr-dg-action-overflow//button[contains(.,"Make Public")]
