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
${artifact_action_xpath}  //clr-dg-action-bar/clr-dropdown/span[contains(@class,'dropdown-toggle')]
${artifact_action_delete_xpath}  //clr-dropdown-menu//div[contains(.,'Delete')]
${artifact_action_copy_xpath}  //clr-dropdown-menu//div[contains(.,'Copy') and @aria-label='retag']
${artifact_achieve_icon}  //artifact-list-tab//clr-datagrid//clr-dg-row[contains(.,'sha256')]//clr-dg-cell[1]//clr-tooltip//clr-icon
${artifact_rows}  //artifact-list-tab//clr-datagrid//clr-dg-row[contains(.,'sha256')]

