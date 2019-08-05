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
${project_member_tag_xpath}  //clr-main-container//project-detail/nav/ul//a[contains(.,'Members')]
${project_member_add_button_xpath}  //project-detail//button[contains(.,'User')]
${project_member_add_username_xpath}  //*[@id='member_name']
${project_member_add_admin_xpath}  /html/body/harbor-app/harbor-shell/clr-main-container/div/div/project-detail/ng-component/div/div[1]/div/div[1]/add-member/clr-modal/div/div[1]/div/div[1]/div/div[2]/form/section/div[2]/div[1]/label
${project_member_add_save_button_xpath}  /html/body/harbor-app/harbor-shell/clr-main-container/div/div/project-detail/ng-component/div/div[1]/div/div[1]/add-member/clr-modal/div/div[1]/div/div[1]/div/div[3]/button[2]
${project_member_search_button_xpath}  //project-detail//hbr-filter/span/clr-icon
${project_member_search_text_xpath}  //project-detail//hbr-filter/span/input
${project_member_add_confirmation_ok_xpath}  //project-detail//add-member//button[2]
${project_member_search_button_xpath2}  //button[contains(.,'New')]
${project_member_add_button_xpath2}  //project-detail//add-member//button[2]
${project_member_guest_radio_checkbox}  //project-detail//form//input[@id='checkrads_guest']
${project_member_delete_button_xpath}  //button[contains(.,'REMOVE')]
