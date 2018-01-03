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

*** Variables ***
${new_name_xpath}  	//hbr-list-replication-rule//button[contains(.,'New')]
${policy_name_xpath}  //*[@id="policy_name"]
${policy_description_xpath}  //*[@id="policy_description"]
${policy_enable_checkbox}  //input[@id='policy_enable']/../label
${policy_endpoint_checkbox}  //input[@id='check_new']/../label
${destination_name_xpath}  //*[@id='destination_name']
${destination_url_xpath}  //*[@id='destination_url']
${destination_username_xpath}  //*[@id='destination_username']
${destination_password_xpath}  //*[@id='destination_password']
${replicaton_save_xpath}  //button[contains(.,'OK')]
