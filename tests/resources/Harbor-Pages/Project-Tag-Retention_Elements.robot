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
${project_tag_retention_add_rule_xpath}  //*[@id='add-rule']
${project_tag_retention_repo_input_xpath}  //*[@id='repos']
${project_tag_retention_param_input_xpath}  //*[@id='param']
${project_tag_retention_tags_input_xpath}  //*[@id='tags']
${project_tag_retention_save_add_button_xpath}  //*[@id='save-add']
${project_tag_retention_template_xpath}  //*[@id='template']
${project_tag_retention_option_always_xpath}  //option[@value='always']
${project_tag_retention_rule_name_xpath}  //ul//span[@class='rule-name ml-5']
${project_tag_retention_edit_schedule_xpath}  //*[@id='editSchedule']
${project_tag_retention_select_policy_xpath}  //*[@id='selectPolicy']
${project_tag_retention_option_daily_xpath}  //option[@value='Daily']
${project_tag_retention_config_save_xpath}  //*[@id='config-save']
${project_tag_retention_schedule_ok_xpath}  //*[@id='schedule-ok']
${project_tag_retention_span_daily_xpath}  //cron-selection//div//span[contains(.,'0 0 0 * * *')]
${project_tag_retention_dry_run_xpath}  //*[@id='dry-run']
${project_tag_retention_refresh_xpath}  //clr-dg-action-bar/button[4]
${project_tag_retention_record_yes_xpath}  //clr-datagrid[contains(.,'Yes')]
${project_tag_retention_list_expand_icon_xpath}  //project-detail/app-tag-feature-integration/tag-retention//clr-datagrid//clr-dg-row//clr-expandable-animation//cds-icon[@class='datagrid-expandable-caret-icon']
${project_tag_retention_run_now_xpath}  //*[@id='run-now']
${project_tag_retention_execute_run_xpath}  //*[@id='execute-run']
${project_tag_retention_record_no_xpath}  //clr-datagrid[contains(.,'No')]
${project_tag_retention_action_button_xpath}  //button[contains(.,'ACTION')]
${project_tag_retention_delete_button_xpath}  //div[contains(@class,'dropdown-menu')]//button[contains(.,'Delete')]
${project_tag_retention_edit_button_xpath}  //div[contains(@class,'dropdown-menu')]//button[contains(.,'Edit')]
${project_tag_retention_modal_title_edit_xpath}  //h3[contains(.,'Edit Tag Retention Rule')]

${project_tag_immutability_scope_input_xpath}  //*[@id='scope-input']
${project_tag_immutability_tag_input_xpath}  //*[@id='tag-input']
${project_tag_immutability_save_add_button_xpath}  //*[@id='add-edit-btn']
