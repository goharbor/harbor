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
${job_service_stop_btn}  //button[.//span[text()=' STOP ']]
${job_service_stop_all_btn}  //button[contains(.,'STOP ALL')]
${job_service_pause_btn}  //button[contains(text(),'PAUSE')]
${job_service_resume_btn}  //button[contains(text(),'RESUME')]
${job_service_refresh_btn}  //clr-datagrid//span[contains(@class,'refresh-btn')]
${job_service_schedules_btn}  //*[@id='schedules']
${job_service_schedules_pause_all_btn}  //button[.//span[text()=' PAUSE ALL ']]
${job_service_schedules_resume_all_btn}  //button[.//span[text()=' RESUME ALL ']]
${job_service_workers_btn}  //button[text()=' Workers ']
