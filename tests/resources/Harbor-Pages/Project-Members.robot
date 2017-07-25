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
Go Into Project
	[Arguments]  ${project}
	Sleep  2
	Click Element  xpath=//*[@id="search_input"]
	Sleep  2
	Input Text  xpath=//*[@id="search_input"]  ${project}
	Sleep  8
	Wait Until Page Contains  ${project}
	Click Element  xpath=//*[@id="results"]/list-project-ro/clr-datagrid/div/div/div/div/div[2]/clr-dg-row[1]/clr-dg-row-master/clr-dg-cell[1]/a
	Sleep  2
	Capture Page Screenshot  gointo_${project}.png

Go Into Project2
	[Arguments]  ${project}
	Sleep  2
	Capture Page Screenshot  gointo1_${project}.png
	# search icon
	Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/project/div/div/div[2]/div[2]/hbr-filter/span/clr-icon/svg
	Sleep  2
	# text search project
	Input Text  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/project/div/div/div[2]/div[2]/hbr-filter/span/input  ${project}
	Sleep  5
	Wait Until Page Contains  ${project}
	Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/project/div/div/list-project/clr-datagrid/div/div/div/div/div[2]/clr-dg-row/clr-dg-row-master/clr-dg-cell[2]/a
	Sleep  3
	Capture Page Screenshot  gointo2_${project}.png
		
Add User To Project Admin
	[Arguments]  ${project}  ${user}
	Go Into Project2
	Sleep  2  
	Click Element  xpath=${project_member_tag_xpath}
	Sleep  1
    Click Element  xpath=${project_member_add_button_xpath}
    Sleep  2	
    Input Text  xpath=${project_member_add_username_xpath}  ${user}
	Sleep  3
    Click Element  xpath=${project_member_add_admin_xpath}
    Click Element  xpath=${project_member_add_save_button_xpath}
	Sleep  4
	
Search Project Member
	[Arguments]  ${project}  ${user}
	Go Into Project  ${project}
	Sleep  2   
	Click Element  xpath=${project_member_tag_xpath}
	Sleep  1	
	Click Element  xpath=${project_member_search_button_xpath}
	Sleep  1	
	Click Element  xpath=${project_member_search_text_xpath}
	Sleep  2
	Wait Until Page Contains  ${user}	
	
Change Project Member Role
	[Arguments]  ${project}  ${user}  ${role}
	Click Element  xpath=//clr-dg-cell//a[contains(.,"${project}")]
    Sleep  2    
	Click Element  xpath=${project_member_tag_xpath}
	Sleep  1	
	Click Element  xpath=//project-detail//clr-dg-row-master[contains(.,"${user}")]//clr-dg-action-overflow
	Sleep  1	
    Click Element  xpath=//project-detail//clr-dg-action-overflow//button[contains(.,"${role}")]
	Sleep  2
	Wait Until Page Contains  ${role}
	
	
	