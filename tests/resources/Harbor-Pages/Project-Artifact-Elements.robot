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
${artifact_action_xpath}  //*[@id='artifact-list-action']
${artifact_action_delete_xpath}  //clr-dropdown-menu//div[contains(.,'Delete')]
${artifact_action_copy_xpath}  //clr-dropdown-menu//div[contains(.,'Copy') and @aria-label='retag']
${artifact_achieve_icon}  //artifact-list-tab//clr-datagrid//clr-dg-row[contains(.,'sha256')]//clr-dg-cell[1]//clr-tooltip//a
${artifact_rows}  //artifact-list-tab//clr-datagrid//clr-dg-row[contains(.,'sha256')]
${archive_rows}  //artifact-list-tab//clr-datagrid//clr-dg-row[contains(.,'sha256')]//clr-dg-cell[1]//clr-tooltip//a
${artifact_list_refresh_btn}  //artifact-list-tab//div//span[@class='refresh-btn']

${artifact_list_spinner}   xpath=//clr-datagrid//clr-spinner
${artifact_tag_component}   xpath=//artifact-tag
${add_tag_button}          xpath=//*[@id='new-tag']
${tag_name_xpath}          xpath=//*[@id='name']
${add_ok_button}           xpath=//*[@id='add-ok']
${delete_tag_button}       xpath=//*[@id='delete-tag']
${dialog_delete_button}    xpath=//clr-modal//button[contains(.,'DELETE')]

${harbor_helm_name}  harbor-helm-1.7.3
${harbor_helm_filename}  harbor-helm-1.7.3.tar.gz
${harbor_helm_version}  1.7.3
${harbor_helm_package}  harbor-1.7.3.tgz

${artifact_list_accessory_btn}  (//clr-dg-row//button)[1]
${artifact_cosign_accessory}  //clr-dg-row//clr-dg-row[./clr-expandable-animation/div/div/div/clr-dg-cell/div[text()=' signature.cosign ']]
${artifact_sbom_accessory}  //clr-dg-row//clr-dg-row[./clr-expandable-animation/div/div/div/clr-dg-cell/div[text()=' subject.accessory ']]
${artifact_cosign_accessory_action_btn}  (//clr-dg-row//clr-dg-row[.//div[text()=' signature.cosign ']]//button)[1]
${artifact_sbom_accessory_action_btn}  (//clr-dg-row//clr-dg-row[.//div[text()=' subject.accessory ']]//button)[1]
${artifact_cosign_cosign_accessory_action_btn}  ${artifact_cosign_accessory}//clr-dg-row//button
${artifact_sbom_cosign_accessory_action_btn}  ${artifact_sbom_accessory}//clr-dg-row//button
${artifact_list_cosign_accessory_btn}  (${artifact_cosign_accessory}//button)[2]
${artifact_list_sbom_accessory_btn}  (${artifact_sbom_accessory}//button)[2]
${copy_digest_btn}  //button[text()=' Copy Digest ']
${delete_accessory_btn}  //button[text()=' Delete ']
${copy_btn}  //button[text()=' COPY ']
${artifact_digest}  //textarea
