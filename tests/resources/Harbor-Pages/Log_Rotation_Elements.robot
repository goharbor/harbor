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
${log_rotation_page_xpath}  //app-clearing-job//nav//a[contains(.,'Log Rotation')]
${keep_records_input}  //*[@id='retentionTime']
${keep_records_unit_select}  //*[@id='expiration-type']
${latest_purge_job_status_xpath}  //app-purge-history//div//clr-dg-row[1]//clr-dg-cell[4]
${latest_purge_job_update_time_xpath}  //app-purge-history//div//clr-dg-row[1]//clr-dg-cell[6]
${purge_job_last_completed_time_xpath}  //app-set-job//div//span[contains(@class,'mr-3')]
${purge_now_btn}  //app-set-job//button[contains(.,'PURGE NOW')]
${log_rotation_schedule_edit_btn}  //*[@id='editSchedule']
${log_rotation_schedule_select}  //*[@id='selectPolicy']
${log_rotation_schedule_save_btn}  //*[@id='config-save']
${log_rotation_schedule_cron_input}  //*[@id='targetCron']