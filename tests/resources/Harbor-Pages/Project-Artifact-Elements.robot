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

# ${artifact_list_accessory_btn}  (//clr-dg-row//button)[1]
# ${artifact_cosign_accessory}  //clr-dg-row//clr-dg-row[.//div[text()=' signature.cosign ']]
# ${artifact_sbom_accessory}  //clr-dg-row//clr-dg-row[.//div[text()=' subject.accessory ']]
# ${artifact_list_cosign_accessory_btn}  xpath=//clr-dg-row[contains(., 'latest')]//button[contains(@class, 'datagrid-expandable-caret-button')]
# ${artifact_list_sbom_accessory_btn}  xpath=//clr-dg-row[contains(., 'latest')]//button[contains(@class, 'datagrid-expandable-caret-button')]
${artifact_sbom_cosign_accessory_action_btn}    xpath=//sub-accessories//clr-dg-row[contains(., 'signature.cosign')]//button[contains(@class, 'datagrid-action-toggle')]
${artifact_cosign_cosign_accessory_action_btn}    xpath=//sub-accessories//clr-dg-row[contains(., 'signature.cosign')]//button[contains(@class, 'datagrid-action-toggle')]
# ${artifact_sbom_accessory_action_btn}    xpath=//clr-dg-row[contains(., 'subject.accessory')]//button[contains(@class, 'datagrid-action-toggle')]
# ${artifact_cosign_accessory_action_btn}    xpath=//clr-dg-row[contains(., 'signature.cosign') and not(ancestor::sub-accessories)]//button[contains(@class, 'datagrid-action-toggle')]
# ${copy_digest_btn}  //button[text()=' Copy Digest ']
# ${delete_accessory_btn}  //button[text()=' Delete ']
# ${copy_btn}  //button[text()=' COPY ']
# ${artifact_digest}  //textarea

# # Level 1: Expand/collapse caret button for the main 'latest' image row
# ${artifact_list_accessory_btn}  xpath=//clr-dg-row[contains(., 'latest')]//button[contains(@class, 'datagrid-expandable-caret-button')]

# # Level 2 (Sub-grid): Row locators for SBOM and independent Cosign signatures
# ${artifact_sbom_accessory}      xpath=//clr-dg-row//clr-dg-row[.//clr-dg-cell[contains(., 'subject.accessory')]]
# ${artifact_cosign_accessory}    xpath=//clr-dg-row//clr-dg-row[.//clr-dg-cell[contains(., 'signature.cosign')]]

# # Level 3 (Nested Sub-grid): Expand caret button nested inside the Level 2 SBOM/Signature row
# ${artifact_list_sbom_accessory_btn}    xpath=//clr-dg-row[.//clr-dg-cell[contains(., 'subject.accessory')]]//button[contains(@class, 'datagrid-expandable-caret-button')]
# ${artifact_list_cosign_accessory_btn}  xpath=//clr-dg-row[.//clr-dg-cell[contains(., 'signature.cosign')]]//button[contains(@class, 'datagrid-expandable-caret-button')]

# # Action Triggers: Ellipsis/Action toggle buttons for SBOM and Signature rows in Level 2
# ${artifact_sbom_accessory_action_btn}    xpath=//clr-dg-row[.//clr-dg-cell[contains(., 'subject.accessory')]]//clr-icon[@shape='ellipsis-vertical' or contains(@class, 'datagrid-action-toggle')]
# ${artifact_cosign_accessory_action_btn}  xpath=//clr-dg-row[.//clr-dg-cell[contains(., 'signature.cosign')]]//clr-icon[@shape='ellipsis-vertical' or contains(@class, 'datagrid-action-toggle')]

# # Contextual Locators: Captures the digest span of the currently active/selected row (removes legacy //textarea dependency)
# ${artifact_digest}  xpath=//clr-dg-row[contains(@class, 'active') or contains(@class, 'selected')]//span[contains(@class, 'digest')]

# Standard UI Button Triggers
${copy_digest_btn}       //button[text()=' Copy Digest ']
${delete_accessory_btn}  //button[text()=' Delete ']
${copy_btn}              //button[text()=' COPY ']

# Level 1: Main row expander icon
${artifact_list_accessory_btn}          xpath=//clr-dg-row[contains(., 'latest')]//button[contains(@class, 'datagrid-expandable-caret-button')]

# Level 2 & 3 Expanders: Explicitly targeting row contents directly to prevent parent-child bleed
${artifact_list_sbom_accessory_btn}    xpath=//clr-dg-cell[contains(., 'subject.accessory')]/preceding-sibling::button or //clr-dg-row[.//clr-dg-cell[contains(., 'subject.accessory')]]/div[contains(@class,'clr-dg-cell')]/button
${artifact_list_cosign_accessory_btn}  xpath=//clr-dg-cell[contains(., 'signature.cosign')]/preceding-sibling::button

# Level 2 Action Triggers: Ellipsis button for SBOM and independent Cosign signature
# 🚨 FIX: We isolate the exact row by targeting the specific cell's adjacent or ancestor trigger structure, avoiding generic //clr-icon.
# ${artifact_sbom_accessory_action_btn}    xpath=//clr-dg-row[div/clr-dg-cell[contains(., 'subject.accessory')]]//button[contains(@class, 'datagrid-action-toggle') or .//clr-icon[@shape='ellipsis-vertical']]
# ${artifact_cosign_accessory_action_btn}  xpath=//clr-dg-row[div/clr-dg-cell[contains(., 'signature.cosign') and not(ancestor::sub-accessories)]]//button[contains(@class, 'datagrid-action-toggle') or .//clr-icon[@shape='ellipsis-vertical']]

# Level 3 Action Triggers: Target action toggles nested inside sub-accessories container exclusively
${artifact_sbom_cosign_accessory_action_btn}    xpath=//sub-accessories//clr-dg-row[contains(., 'signature.cosign')]//button[contains(@class, 'datagrid-action-toggle') or .//clr-icon[@shape='ellipsis-vertical']]
${artifact_cosign_cosign_accessory_action_btn}  xpath=//sub-accessories//clr-dg-row[contains(., 'signature.cosign')]//button[contains(@class, 'datagrid-action-toggle') or .//clr-icon[@shape='ellipsis-vertical']]

# Contextual Digest Picker
${artifact_digest}                      xpath=//clr-dg-row[contains(@class, 'active') or contains(@class, 'selected') or @aria-selected='true']//span[contains(@class, 'digest') or contains(., 'sha256')]

${artifact_sbom_accessory_action_btn}    xpath=//clr-dg-cell[contains(., 'subject.accessory')]/..//button[contains(@class, 'toggle') or .//clr-icon]
${artifact_cosign_accessory_action_btn}  xpath=//clr-dg-cell[contains(., 'signature.cosign') and not(ancestor::sub-accessories)]/..//button[contains(@class, 'toggle') or .//clr-icon]