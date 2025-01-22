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
Resource  ../../resources/Util.robot
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${ip}
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Distribution CRUD
    ${d}=    Get Current Date    result_format=%m%s
    ${name}=  Set Variable  distribution${d}
    ${endpoint}=  Set Variable  https://32.1.1.2
    ${endpoint_new}=  Set Variable  https://10.65.65.42
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Distribution  Dragonfly  ${name}  ${endpoint}  ${DRAGONFLY_AUTH_TOKEN}
    Edit A Distribution  ${name}  ${endpoint}  new_endpoint=${endpoint_new}
    Delete A Distribution  ${name}  ${endpoint_new}
    Close Browser

Test Case - P2P Preheat Policy CRUD
    ${d}=    Get Current Date    result_format=%m%s
    ${project_name}=  Set Variable  project_p2p${d}
    ${dist_name}=  Set Variable  distribution${d}
    ${endpoint}=  Set Variable  https://20.76.1.2
    ${policy_name}=  Set Variable  policy${d}
    ${repo}=  Set Variable  alpine
    ${repo_new}=  Set Variable  redis*
    ${tag}=  Set Variable  v1.0
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Distribution  Dragonfly  ${dist_name}  ${endpoint}  ${DRAGONFLY_AUTH_TOKEN}
    Create An New Project And Go Into Project  ${project_name}
    Create An New P2P Preheat Policy  ${policy_name}  ${dist_name}  ${repo}  ${tag}
    Edit A P2P Preheat Policy  ${policy_name}  ${repo_new}
    Delete A Distribution  ${dist_name}  ${endpoint}  deletable=${false}
    Go Into Project  ${project_name}  has_image=${false}
    Delete A P2P Preheat Policy  ${policy_name}
    Delete A Distribution  ${dist_name}  ${endpoint}
    Close Browser

Test Case - P2P Preheat By Manual
    [Tags]  p2p_preheat_by_manual  need_distribution_endpoint
    ${d}=    Get Current Date    result_format=%m%s
    ${project_name}=  Set Variable  project_p2p${d}
    ${dist_name}=  Set Variable  distribution${d}
    ${policy_name}=  Set Variable  policy${d}
    ${image1}=  Set Variable  busybox
    ${image2}=  Set Variable  hello-world
    ${tag1}=  Set Variable  latest
    ${tag2}=  Set Variable  stable
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Distribution  Dragonfly  ${dist_name}  ${DISTRIBUTION_ENDPOINT}  ${DRAGONFLY_AUTH_TOKEN}
    Create An New Project And Go Into Project  ${project_name}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image1}  ${tag1}  ${tag1}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image1}  ${tag2}  ${tag2}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image2}  ${tag1}  ${tag1}
    Create An New P2P Preheat Policy  ${policy_name}  ${dist_name}  ${image1}  ${tag1}
    ${contain}  Create List  ${project_name}/${image1}:${tag1}
    ${not_contain}  Create List  ${project_name}/${image1}:${tag2}  ${project_name}/${image2}:${tag1}
    Retry Action Keyword  Execute P2P Preheat And Verify  ${project_name}  ${policy_name}  ${contain}  ${not_contain}
    Delete A P2P Preheat Policy  ${policy_name}
    Delete A Distribution  ${dist_name}  ${DISTRIBUTION_ENDPOINT}
    Close Browser

Test Case - P2P Preheat By Event
    [Tags]  p2p_preheat_by_event  need_distribution_endpoint
    ${d}=    Get Current Date    result_format=%m%s
    ${project_name}=  Set Variable  project_p2p${d}
    ${dist_name}=  Set Variable  distribution${d}
    ${policy_name}=  Set Variable  policy${d}
    ${image1}=  Set Variable  busybox
    ${image2}=  Set Variable  hello-world
    ${tag1}=  Set Variable  latest
    ${tag2}=  Set Variable  stable
    ${label}=  Set Variable  p2p_preheat
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Distribution  Dragonfly  ${dist_name}  ${DISTRIBUTION_ENDPOINT}  ${DRAGONFLY_AUTH_TOKEN}
    Create An New Project And Go Into Project  ${project_name}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image1}  ${tag1}  ${tag1}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image1}  ${tag2}  ${tag2}
    Create An New P2P Preheat Policy  ${policy_name}  ${dist_name}  **  **  Event based
    Retry Double Keywords When Error  Select P2P Preheat Policy  ${policy_name}  Wait Until Element Is Visible  ${p2p_execution_header}
    # Artifact is pushed event
    Retry Action Keyword  Verify Artifact Is Pushed Event  ${project_name}  ${policy_name}  ${image2}  ${tag1}
    # Artifact is scanned event
    Retry Action Keyword  Verify Artifact Is Scanned Event  ${project_name}  ${policy_name}  ${image1}  ${tag1}
    # Artifact is labeled event
    Retry Action Keyword  Verify Artifact Is Labeled Event  ${project_name}  ${policy_name}  ${image1}  ${tag2}  ${label}
    Delete A P2P Preheat Policy  ${policy_name}
    Delete A Distribution  ${dist_name}  ${DISTRIBUTION_ENDPOINT}
    Close Browser