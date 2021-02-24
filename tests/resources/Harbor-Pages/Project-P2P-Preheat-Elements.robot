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

*** Variables ***
${project_p2p_preheat_tag_xpath}  //clr-main-container//project-detail/clr-tabs//a[contains(.,'P2P Preheat')]
${p2p_preheat_new_policy_btn_id}  //*[@id='new-policy']
${p2p_preheat_provider_select_id}  //*[@id='provider']
${p2p_preheat_name_input_id}  //*[@id='name']
${p2p_preheat_repoinput_id}  //*[@id='repo']
${p2p_preheat_tag_input_id}  //*[@id='tag']
${p2p_preheat_trigger_type_select_id}  //*[@id='trigger-type']
${p2p_preheat_add_save_btn_id}  //add-p2p-policy//*[@id='new-policy']
${p2p_preheat_edit_save_btn_id}  //*[@id='edit-policy-save']
${p2p_preheat_action_btn_id}  //*[@id='action-policy']
${p2p_preheat_del_btn_id}  //*[@id='delete-policy']
${p2p_preheat_edit_btn_id}  //*[@id='edit-policy']
${p2p_execution_header}  //clr-main-container//project-detail//ng-component//h4[contains(.,'Executions')]



