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
${sign_up_for_an_account_xpath}  /html/body/harbor-app/sign-in/clr-main-container/div/form/div[1]/a
${sign_up_button_xpath}  //a[@class='signup']
${username_xpath}  //*[@id='username']
${email_xpath}  //*[@id='email']
${realname_xpath}  //*[@id='realname']
${newPassword_xpath}  //*[@id='newPassword']
${confirmPassword_xpath}  //*[@id='confirmPassword']
${comment_xpath}  //*[@id='comment']
${signup_xpath}  //*[@id='sign-up']
${search_input}  //*[@id='search_input']
${login_btn}  //*[@id='log_in']
${harbor_span_title}  //span[contains(., 'Harbor')]
${login_name}  //*[@id='login_username']
${login_pwd}  //*[@id='login_password']
${header_user}  //clr-dropdown//button[@class='nav-text dropdown-toggle']//span
${about_btn}  //clr-dropdown-menu/a[contains(.,'About')]
${header}  xpath=//clr-header[contains(@class,'header-5')]
${color_theme_light}  //span[contains(.,'LIGHT')]
${close_btn}  //button[contains(.,'CLOSE')]
${more_info_link_xpath}  //sign-in//div//a[contains(.,'More info...')]
${event_log_xpath}  //hbr-operation-model//a[contains(., 'EVENT LOG')]
${latest_download_cve_csv_file_xpath}  //section/div[1]//clr-icon[@shape='download']
${latest_cve_csv_file_name_xpath}  //section/div[1]//span[contains(., 'csv_file')]