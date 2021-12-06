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
${new_sys_robot_account_btn}                 //system-robot-accounts//button/span/span[contains(.,'NEW ROBOT ACCOUNT')]
${sys_robot_account_name_input}              //*[@id='name']
${sys_robot_account_expiration_type_select}  //*[@id='expiration-type']
${sys_robot_account_expiration_input}        //*[@id='robotTokenExpiration']
${sys_robot_account_description_textarea}    //*[@id='description']
${sys_robot_account_coverall_chb_input}  xpath=//input[@id='coverAll']
${sys_robot_account_coverall_chb}            //clr-checkbox-wrapper[contains(@class, 'clr-checkbox-wrapper')]/label[contains(@for, 'coverAll')]
${sys_robot_account_permission_list_btn}     //form/section//clr-dropdown/button
${save_sys_robot_account_btn}                //*[@id='system-robot-save']
${save_sys_robot_export_to_file_btn}         //section//button
${save_sys_robot_project_filter_chb}         //clr-dg-string-filter/clr-dg-filter//cds-icon
${save_sys_robot_project_filter_input}       //input[contains(@name, 'search')]
${save_sys_robot_project_filter_close_btn}   //button/cds-icon[contains(@title, 'Close')]
${save_sys_robot_project_paste_icon}         //hbr-copy-input//clr-icon
