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
${head_admin_xpath}  xpath=//clr-dropdown//button//clr-icon[@shape='user']
${change_password_xpath}  xpath=//clr-main-container//clr-dropdown//a[2]
${user_profile_xpath}  xpath=//clr-main-container//clr-dropdown//a[1]
${old_password_xpath}  xpath=//*[@id='oldPassword']
${new_password_xpath}  xpath=//*[@id='newPassword']
${renew_password_xpath}  xpath=//*[@id='reNewPassword']
${change_password_confirm_btn_xpath}  xpath=//password-setting/clr-modal//button[2]
${user_profile_confirm_btn_xpath}  xpath=//account-settings-modal/clr-modal//button[2]
${sign_in_title_xpath}  xpath=//sign-in//form//*[@class='title']
${account_settings_comments_xpath}  xpath=//*[@id='account_settings_comments']
${about_xpath}  xpath=//clr-dropdown-menu//a[contains(.,'About')]
${license_xpath}  xpath=//about-dialog//div//p//a[contains(.,'Open Source/Third Party License')]