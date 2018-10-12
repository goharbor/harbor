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
${project_create_xpath}  //clr-dg-action-bar//button[contains(.,'New')]
${self_reg_xpath}  //input[@id="clr-checkbox-selfReg"]
${test_ldap_xpath}  /html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/div/button[3]
${config_save_button_xpath}  //config//div/button[contains(.,'SAVE')]
${configuration_xpath}  //clr-vertical-nav-group-children/a[contains(.,'Configuration')]
${system_config_xpath}  //*[@id="config-system"]
${garbage_collection_xpath}  //*[@id="config-gc"]
${gc_now_xpath}  //*[@id="gc"]/gc-config/button
${gc_log_details_xpath}  /html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/section[6]/gc-config/clr-datagrid/div/div/div/clr-dg-table-wrapper/div[2]/clr-dg-row/div/clr-dg-cell[6]/a