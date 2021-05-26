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
Documentation  Harbor BATs
Library  ../../apitests/python/testutils.py
Library  ../../apitests/python/library/oras.py
Library  ../../apitests/python/library/singularity.py
Resource  ../../resources/Util.robot
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${ip}
${SSH_USER}  root
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Customize Look
    [tags]  look
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Retry Wait Element  //span[contains(., 'Harbor product name')]
    Retry Element Click  ${header_user}
    Retry Element Click  ${about_btn}
    Retry Wait Element  //p[contains(., 'test customize look for harbor')]
    Retry Element Click  ${close_btn}
    ${style}=   Get Element Attribute  ${header}  style
    Log All  ${style}
    Should Contain  ${style}  background-color: red
    Retry Element Click  ${color_theme_light}
    Sleep  2
    ${style}=   Get Element Attribute  ${header}  style
    Log All  ${style}
    Should Contain  ${style}  background-color: yellow
    Close Browser