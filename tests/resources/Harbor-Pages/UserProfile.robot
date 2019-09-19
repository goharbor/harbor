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
Resource  ../../resources/Util.robot

*** Variables ***

*** Keywords ***
Change Password
    [Arguments]  ${cur_pw}  ${new_pw}
    Retry Element Click  ${head_admin_xpath}
    Retry Element Click  ${change_password_xpath}
    Retry Text Input  ${old_password_xpath}  ${cur_pw}
    Retry Text Input  ${new_password_xpath}   ${new_pw}
    Retry Text Input  ${renew_password_xpath}  ${new_pw}
    Retry Element Click  ${change_password_confirm_btn_xpath}
    Retry Element Click  xpath=${log_xpath}
    Sleep  1

Update User Comment
    [Arguments]  ${new_comment}
    Retry Element Click  ${head_admin_xpath}
    Retry Element Click  ${user_profile_xpath}
    Retry Text Input  ${account_settings_comments_xpath}  ${new_comment}
    Retry Element Click  ${user_profile_confirm_btn_xpath}
    Sleep  2

Logout Harbor
    Retry Element Click  ${head_admin_xpath}
    Retry Link Click  Log Out
    Capture Page Screenshot  Logout.png
    Sleep  2
    Wait Until Keyword Succeeds  5x  1  Page Should Contain Element  ${sign_in_title_xpath}