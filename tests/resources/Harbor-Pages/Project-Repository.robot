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

*** Keywords ***
View Repo Scan Details
    Retry Element Click  xpath=${first_repo_xpath}
    Capture Page Screenshot
    Retry Wait Until Page Contains  unknown
    Retry Wait Until Page Contains  high
    Retry Wait Until Page Contains  medium
    Retry Wait Until Page Contains  CVE
    Retry Element Click  xpath=${build_history_btn}
    Retry Wait Until Page Contains Element  xpath=${build_history_data}

View Scan Error Log
    Retry Wait Until Page Contains  View Log
    Retry Element Click  xpath=${view_log_xpath}
    Capture Page Screenshot  viewlog.png


