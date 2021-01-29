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
Switch To P2P Preheat
    Retry Element Click  xpath=${project_p2p_preheat__tag_xpath}

Select Distribution For P2P Preheat
    [Arguments]    ${provider}
    Retry Element Click    ${p2p_preheat_provider_select_id}
    Retry Element Click    ${p2p_preheat_provider_select_id}//option[contains(.,'${provider}')]

Select P2P Preheat Policy
    [Arguments]    ${name}
    Retry Element Click    //clr-dg-row[contains(.,'${name}')]//clr-radio-wrapper/label

P2P Preheat Policy Exist
    [Arguments]  ${name}  ${repo}=${null}
    ${policy_row_xpath}=  Set Variable If  '${repo}'=='${null}'  //clr-dg-row[contains(.,'${name}')]  //clr-dg-row[contains(.,'${name}') and contains(.,'${repo}')]
    Retry Wait Until Page Contains Element  ${policy_row_xpath}

P2P Preheat Policy Not Exist
    [Arguments]  ${name}
    Retry Wait Until Page Not Contains Element  //clr-dg-row[contains(.,'${name}')]

Create An New P2P Preheat Policy
    [Arguments]    ${policy_name}  ${dist_name}  ${repo}  ${tag}  ${trigger_type}=${null}
    Switch To P2P Preheat
    Retry Element Click  ${p2p_preheat_new_policy_btn_id}
    Select Distribution For P2P Preheat  ${dist_name}
    Retry Text Input  ${p2p_preheat_name_input_id}  ${policy_name}
    Retry Text Input  ${p2p_preheat_repoinput_id}  ${repo}
    Retry Text Input  ${p2p_preheat_tag_input_id}  ${tag}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_add_save_btn_id}  Retry Wait Until Page Not Contains Element  xpath=${p2p_preheat_add_save_btn_id}
    P2P Preheat Policy Exist  ${policy_name}

Edit A P2P Preheat Policy
    [Arguments]    ${name}  ${repo}  ${trigger_type}=${null}
    Switch To P2P Preheat
    Retry Double Keywords When Error  Select P2P Preheat Policy   ${name}  Wait Until Element Is Visible  ${p2p_execution_header}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_action_btn_id}  Wait Until Element Is Visible And Enabled  ${p2p_preheat_edit_btn_id}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_edit_btn_id}  Wait Until Element Is Visible And Enabled  ${p2p_preheat_name_input_id}
    Retry Text Input  ${p2p_preheat_repoinput_id}  ${repo}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_edit_save_btn_id}  Retry Wait Until Page Not Contains Element  xpath=${p2p_preheat_edit_save_btn_id}
    P2P Preheat Policy Exist  ${name}  repo=${repo}

Delete A P2P Preheat Policy
    [Arguments]    ${name}
    Switch To P2P Preheat
    Retry Double Keywords When Error  Select P2P Preheat Policy   ${name}  Wait Until Element Is Visible  ${p2p_execution_header}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_action_btn_id}  Wait Until Element Is Visible And Enabled  ${p2p_preheat_del_btn_id}
    Retry Double Keywords When Error  Retry Element Click  ${p2p_preheat_del_btn_id}  Wait Until Element Is Visible And Enabled  ${delete_confirm_btn}
    Retry Double Keywords When Error  Retry Element Click  ${delete_confirm_btn}  Retry Wait Until Page Not Contains Element  ${delete_confirm_btn}
    P2P Preheat Policy Not Exist  ${name}
