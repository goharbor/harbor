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
Documentation  This resource provides helper functions for docker operations
Library  OperatingSystem
Library  Process

*** Keywords ***
CNAB Push Bundle
    [Arguments]  ${ip}  ${user}  ${pwd}  ${target}  ${bundle_file}  ${registry}  ${namespace}  ${index1}  ${index2}
    Wait Unitl Command Success  ./tests/robot-cases/Group0-Util/cnab_push_bundle.sh ${ip} ${user} ${pwd} ${target} ${bundle_file} ${registry} ${namespace} ${index1} ${index2}

Prepare Cnab Push Test Data
    [Arguments]  ${ip}  ${user}  ${pwd}  ${project}  ${index1_image1}  ${index1_image2}  ${index2_image1}  ${index2_image2}  ${image_tag}=latest
    ${index1} =  Set Variable  index1
    ${index2} =  Set Variable  index2
    ${index1_tag} =  Set Variable  latest
    ${index2_tag} =  Set Variable  latest
    Push image  ${ip}  ${user}  ${pwd}  ${project}  ${index1_image1}
    Push image  ${ip}  ${user}  ${pwd}  ${project}  ${index1_image2}
    Push image  ${ip}  ${user}  ${pwd}  ${project}  ${index2_image1}
    Push image  ${ip}  ${user}  ${pwd}  ${project}  ${index2_image2}
    Go Into Project  ${project}
    Wait Until Page Contains  ${project}/${index1_image1}
    Wait Until Page Contains  ${project}/${index1_image2}
    Wait Until Page Contains  ${project}/${index2_image1}
    Wait Until Page Contains  ${project}/${index2_image2}

    Docker Push Index  ${ip}  ${user}  ${pwd}  ${ip}/${project}/${index1}:${index1_tag}  ${ip}/${project}/${index1_image1}:${image_tag}  ${ip}/${project}/${index1_image2}:${image_tag}
    Docker Push Index  ${ip}  ${user}  ${pwd}  ${ip}/${project}/${index2}:${index2_tag}  ${ip}/${project}/${index2_image1}:${image_tag}  ${ip}/${project}/${index2_image2}:${image_tag}
    Go Into Project  ${project}
    Wait Until Page Contains  ${index1}
    Wait Until Page Contains  ${index2}
    [Return]  ${index1}:${index1_tag}  ${index2}:${index2_tag}
