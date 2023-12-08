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
${gc_page_xpath}  //clr-main-container//clr-vertical-nav-group//span[contains(.,'Clean Up')]
${gc_now_button}  //*[@id='gc-now']
${dry_run_button}  //*[@id='gc-dry-run']
${checkbox_delete_untagged_artifacts}  //gc-config//clr-toggle-wrapper/label[contains(@class,'clr-control-label') and contains(@for,'delete_untagged')]
${latest_job_id_xpath}  //clr-datagrid//div//clr-dg-row[1]//clr-dg-cell[1]
${gc_schedule_edit_btn}  //*[@id='editSchedule']
${gc_schedule_select}  //*[@id='selectPolicy']
${gc_schedule_save_btn}  //*[@id='config-save']
${gc_latest_execution_id}  //clr-dg-row[1]//clr-dg-cell[1]
${gc_workers_select}  //*[@id='workers']
