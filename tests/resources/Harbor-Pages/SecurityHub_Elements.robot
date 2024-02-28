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
${total_vulnerabilities_xpath}  (//div[@class='card'][1]//div[contains(@class, 'clr-col-9')])
${security_hub_search_btn}  //button[@id='search']
${top5_most_dangerous_artifacts_xpath}  //div[@class='card'][2]//div[contains(@class,'card-block')]//div[contains(@class,'clr-row')]
${top5_most_dangerous_cves_xpath}  //div[@class='card'][3]//div[contains(@class,'card-block')]//div[contains(@class,'clr-row')]
${add_search_criteria_icon}  //form//clr-icon[@shape='plus-circle']
${add_search_criteria_icon_disabled}  //form//clr-icon[@shape='plus-circle' and contains(@class,'disabled')]
${remove_search_criteria_icon}  //form//clr-icon[@shape='minus-circle']
${remove_search_criteria_icon_disabled}  //form//clr-icon[@shape='minus-circle' and contains(@class,'disabled')]
${vulnerabilities_count_xpath}  //clr-dg-footer//div[contains(@class,'datagrid-footer-description')]//span
${vulnerabilities_filter_select}  (//form//div[@class='clr-select-wrapper']//select)
${vulnerabilities_filter_input}  (//form[contains(@class,'clr-form')]//input)
${vulnerabilities_datagrid_row}  //clr-datagrid//clr-dg-row
${vulnerabilities_filter_label_xpath}  //form//clr-dropdown[contains(@class,'dropdown')]
