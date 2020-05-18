
// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${ip}
${SSH_USER}  root
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Scan Schedule Job
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%M
    Log To Console  ${d}
    ${project_name}=  Set Variable  scan_schedule_proj
    ${image}=  Set Variable  redis
    ${tag}=  Set Variable  latest
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  ${project_name}
    Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image}:${tag}
    Go Into Repo  ${project_name}/${image}
    Retry Wait Until Page Contains Element  ${not_scanned_icon}
    Switch To Vulnerability Page
    ${flag}=  Set Variable  ${false}
    :FOR    ${i}    IN RANGE    999999
    \    ${minite}=  Get Current Date  result_format=%M
    \    ${left} =  Evaluate 	${minite}%10
    \    ${d} =  Convert To Integer  ${left}
    \    Log To Console    ${i}/${d}
    \    Run Keyword If  ${d} <= 3  Run Keywords  Set Scan Schedule  custom  value=* */10 * * * *  AND  Set Suite Variable  ${flag}  ${true}
    \    Sleep  55
    \    Exit For Loop If    '${flag}' == '${true}'
    # After scan custom schedule is set, image should stay in unscanned status.
    Sleep  360
    Go Into Project  ${project_name}
    Go Into Repo  ${project_name}/${image}
    Retry Wait Until Page Contains Element  ${not_scanned_icon}
    Sleep  360
    Scan Result Should Display In List Row  ${tag}
    View Repo Scan Details  Critical  High  Medium


