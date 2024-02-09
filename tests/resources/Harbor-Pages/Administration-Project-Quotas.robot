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
Switch to Project Quotas Tag
    Retry Element Click  xpath=${administration_project_quotas_tag_xpath}
    Sleep  1

Check Project Quota Sorting
    [Arguments]   ${proj1}  ${proj2}
    # check project quota sorting in ascending order
    Retry Element Click  xpath=${sort_used_storage_button}
    Retry Wait Element Visible  //div[@class='datagrid-table']//clr-dg-row[2]//clr-dg-cell[1]//a[contains(text(), '${proj1}')]
    Retry Wait Element Visible  //div[@class='datagrid-table']//clr-dg-row[3]//clr-dg-cell[1]//a[contains(text(), '${proj2}')]
    # check project quota sorting in descending order
    Retry Element Click  xpath=${sort_used_storage_button}
    Retry Wait Element Visible  //div[@class='datagrid-table']//clr-dg-row[1]//clr-dg-cell[1]//a[contains(text(), '${proj2}')]
    Retry Wait Element Visible  //div[@class='datagrid-table']//clr-dg-row[2]//clr-dg-cell[1]//a[contains(text(), '${proj1}')]
