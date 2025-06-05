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
${first_repo_xpath}  //clr-dg-row//clr-dg-cell[1]//a
${first_cve_xpath}  //clr-dg-row[1]//clr-dg-cell//a
${view_log_xpath}  //clr-dg-row//clr-dg-cell//a[contains(.,'View Log')]
${build_history_btn}  //button[contains(.,'Build History')]
${build_history_data}  //clr-dg-row
${push_image_command_btn}  //hbr-push-image-button//button
${scan_artifact_btn}  //button[@id='scan-btn']
${stop_scan_artifact_btn}  //button[@id='stop-scan']
${scan_stopped_label}  //span[normalize-space()='Scan stopped']
${gen_sbom_stopped_label}  //span[normalize-space()='Generation stopped']
${gen_artifact_sbom_btn}  //button[@id='generate-sbom-btn']
${stop_gen_artifact_sbom_btn}  //button[@id='stop-sbom-btn']
${refresh_repositories_xpath}  //hbr-repository-gridview//span[contains(@class,'refresh-btn')]