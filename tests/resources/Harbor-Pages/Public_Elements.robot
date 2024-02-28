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
Documentation  This resource provides any keywords related to public

*** Variables ***
${delete_btn}  //clr-modal//button[contains(.,'DELETE')]
${delete_btn_2}  //button[contains(.,'Delete')]
${default_scanner_info_close_icon}  /html/body/harbor-app/harbor-shell/clr-main-container/div[1]/div[3]/clr-icon
${back_to_home_link}  /html/body/harbor-app/harbor-shell/clr-main-container/div[2]/div/search-result/div/div[2]/a
${select_all_project_box}  //label[contains(@class,'clr-control-label') and contains(@for, 'clr-dg-select-all-clr-id-75')]
${export_cve_btn}  //button[contains(.,'Export CVEs')]
${export_cve_filter_repo_input}  //*[@id='repo']
${export_cve_filter_tag_input}  //*[@id='tag']
${export_cve_filter_cveid_input}  //*[@id='ids']
${export_btn}  //clr-modal//button[contains(.,'EXPORT')]
