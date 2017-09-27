# Copyright 2016-2017 VMware, Inc. All Rights Reserved.
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
${HARBOR_VERSION}  v1.1.1

*** Keywords ***
Create An New Rule With New Endpoint
    [Arguments]  ${policy_name}  ${policy_description}  ${destination_name}  ${destination_url}  ${destination_username}  ${destination_password}
    Click element  xpath=${new_name_xpath}
    Sleep  2
    	
    Input Text  xpath=${policy_name_xpath}  ${policy_name}
    Input Text  xpath=${policy_description_xpath}   ${policy_description}

    Click element  xpath=${policy_enable_checkbox}
    Click element  xpath=${policy_endpoint_checkbox}

    Input text  xpath=${destination_name_xpath}  ${destination_name}
    Input text  xpath=${destination_url_xpath}  ${destination_url}
    Input text  xpath=${destination_username_xpath}  ${destination_username}
    Input text  xpath=${destination_password_xpath}  ${destination_password}
    Click element  xpath=${replicaton_save_xpath}
    Sleep  5
    Capture Page Screenshot  rule_${policy_name}.png
    Wait Until Page Contains  ${policy_name}
    Wait Until Page Contains  ${policy_description}
    Wait Until Page Contains  ${destination_name}