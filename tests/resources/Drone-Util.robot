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
Documentation  This resource contains any keywords related to using the Drone CI Build System

*** Keywords ***
Get State Of Drone Build
    [Arguments]  ${num}
    Return From Keyword If  '${num}' == '0'  local
    ${out}=  Run  drone build info vmware/vic ${num}
    ${lines}=  Split To Lines  ${out}
    [Return]  @{lines}[2]

Get Title of Drone Build
    [Arguments]  ${num}
    Return From Keyword If  '${num}' == '0'  local
    ${out}=  Run  drone build info vmware/vic ${num}
    ${lines}=  Split To Lines  ${out}
    [Return]  @{lines}[-1]
