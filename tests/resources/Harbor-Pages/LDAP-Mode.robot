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
Switch To Configuration Authentication
    Sleep  1
    Retry Element Click  xpath=${configuration_xpath}
    Retry Element Click  xpath=${configuration_authentication_tabsheet_id}

Set LDAP Group Admin DN
    [Arguments]   ${group_dn}
    Switch To Configuration Authentication
    Retry Text Input  ${cfg_auth_ldap_group_admin_dn}  ${group_dn}
    Retry Element Click  ${config_auth_save_button_xpath}

Ldap User Should Not See Change Password
    Retry Element Click  //clr-header//clr-dropdown[2]//button
    Sleep  2
    Page Should Not Contain  Password


