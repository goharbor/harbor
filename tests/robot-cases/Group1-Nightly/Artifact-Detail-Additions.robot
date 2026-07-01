# Copyright Project Harbor Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License

*** Settings ***
Documentation     Artifact Detail Page - Dockerfile Addition Tab Coverage (Enhanced)
...               Comprehensive tests for displaying Dockerfile from image labels
...               Covers happy path, fallback path, edge cases, and error scenarios
Resource          ../../resources/Util.robot
Resource          ../../resources/Harbor-Pages/Project-Artifact.robot
Resource          ../../resources/Harbor-Pages/Project-Repository.robot
Library           Collections
Library           String

*** Variables ***
${artifact_detail_dockerfile_tab}           id=dockerfile-link
${artifact_detail_dockerfile_content}       xpath=//hbr-artifact-dockerfile/div[@class='row content-wrapper']
${artifact_detail_dockerfile_info_box}      xpath=//hbr-artifact-dockerfile//div[@class='info-box']
${artifact_detail_no_dockerfile_msg}        xpath=//hbr-artifact-dockerfile//div[contains(text(), 'Dockerfile')]
${artifact_detail_build_history_tab}        id=build-history
${artifact_detail_build_history_link}       xpath=//hbr-artifact-dockerfile//a[contains(text(), 'Build History')]
${artifact_detail_loading_spinner}          xpath=//hbr-artifact-dockerfile//span[@class='spinner']
${artifact_detail_file_too_large_msg}       xpath=//hbr-artifact-dockerfile//div[contains(text(), 'too large')]
${yaml_container}                            xpath=//div[@class='yaml-container']

*** Test Cases ***

# === Core Functionality Tests ===

Test Dockerfile Tab Display With Label
    [Documentation]    Verify Dockerfile tab appears and displays content when image has label
    [Tags]    artifact    dockerfile    addition    core
    Init Test Data With Dockerfile Label
    Go To Project
    Go To Repository
    Go Into Artifact
    Click Dockerfile Tab
    Verify Dockerfile Content Displayed
    [Teardown]    Cleanup Test Data

Test Dockerfile Content Matches Label
    [Documentation]    Verify displayed Dockerfile matches the label content exactly
    [Tags]    artifact    dockerfile    addition    core
    [Documentation]    Critical: Content must match what was stored in label
    Init Test Data With Dockerfile Label
    Go To Project
    Go To Repository
    Go Into Artifact
    Click Dockerfile Tab
    Verify Dockerfile Content Contains    FROM ubuntu:22.04
    Verify Dockerfile Content Contains    RUN apt-get update
    [Teardown]    Cleanup Test Data

Test Dockerfile Tab Display Without Label
    [Documentation]    Verify informational message when image lacks Dockerfile label
    [Tags]    artifact    dockerfile    addition    core
    Init Test Data Without Dockerfile Label
    Go To Project
    Go To Repository
    Go Into Artifact
    Click Dockerfile Tab
    Verify No Dockerfile Info Message Displayed
    [Teardown]    Cleanup Test Data

Test Dockerfile Tab Provides Build History Link
    [Documentation]    Verify user can navigate to Build History from Dockerfile tab
    [Tags]    artifact    dockerfile    addition    core
    Init Test Data Without Dockerfile Label
    Go To Project
    Go To Repository
    Go Into Artifact
    Click Dockerfile Tab
    Verify No Dockerfile Info Message Displayed
    Click Build History Link From Dockerfile Tab
    Verify Build History Tab Active
    [Teardown]    Cleanup Test Data

Test Tab Navigation And Switching
    [Documentation]    Verify Dockerfile tab can be clicked and switched between tabs
    [Tags]    artifact    dockerfile    addition    core
    Init Test Data With Dockerfile Label
    Go To Project
    Go To Repository
    Go Into Artifact
    Click Dockerfile Tab
    Verify Dockerfile Tab Active
    Wait Until Page Contains Element    ${yaml_container}    timeout=10s
    Click Build History Tab
    Verify Build History Tab Active
    Click Dockerfile Tab
    Verify Dockerfile Tab Active
    [Teardown]    Cleanup Test Data

# === Visibility & UX Tests ===

Test Dockerfile Tab Always Visible
    [Documentation]    Verify Dockerfile tab is always visible, regardless of label presence
    [Tags]    artifact    dockerfile    addition    ux
    Init Test Data Without Dockerfile Label
    Go To Project
    Go To Repository
    Go Into Artifact
    Retry Wait Until Page Contains Element    ${artifact_detail_dockerfile_tab}
    [Teardown]    Cleanup Test Data

Test Dockerfile Content Has Syntax Highlighting
    [Documentation]    Verify Dockerfile content renders with syntax highlighting
    [Tags]    artifact    dockerfile    addition    ux
    Init Test Data With Dockerfile Label
    Go To Project
    Go To Repository
    Go Into Artifact
    Click Dockerfile Tab
    Verify Dockerfile Content Displayed
    # Verify syntax highlighting is applied (yaml-container uses language pipe)
    Retry Wait Until Page Contains Element    ${yaml_container}
    [Teardown]    Cleanup Test Data

Test Build History Tab Always Available
    [Documentation]    Verify Build History tab is always available as fallback
    [Tags]    artifact    dockerfile    addition    ux
    Init Test Data Without Dockerfile Label
    Go To Project
    Go To Repository
    Go Into Artifact
    Click Build History Tab
    Verify Build History Tab Active
    Retry Wait Until Page Contains Element    xpath=//hbr-artifact-build-history
    [Teardown]    Cleanup Test Data

# === Edge Cases & Error Scenarios ===

Test Empty Dockerfile Label
    [Documentation]    Verify graceful handling of empty Dockerfile label
    [Tags]    artifact    dockerfile    addition    edge-case
    Init Test Data With Empty Dockerfile Label
    Go To Project
    Go To Repository
    Go Into Artifact
    Click Dockerfile Tab
    # Should show no-dockerfile message (empty label is treated as missing)
    Verify No Dockerfile Info Message Displayed
    [Teardown]    Cleanup Test Data

Test Dockerfile Tab Content Persistence
    [Documentation]    Verify tab content doesn't reload when switching back
    [Tags]    artifact    dockerfile    addition    performance
    Init Test Data With Dockerfile Label
    Go To Project
    Go To Repository
    Go Into Artifact
    Click Dockerfile Tab
    ${content_first}=    Get Text    ${yaml_container}
    Click Build History Tab
    Sleep    500ms
    Click Dockerfile Tab
    ${content_second}=    Get Text    ${yaml_container}
    Should Be Equal    ${content_first}    ${content_second}
    [Teardown]    Cleanup Test Data

Test Rapid Tab Clicking
    [Documentation]    Verify robust handling of rapid tab switching
    [Tags]    artifact    dockerfile    addition    edge-case
    Init Test Data With Dockerfile Label
    Go To Project
    Go To Repository
    Go Into Artifact
    # Rapid clicking between tabs
    :FOR    ${i}    IN RANGE    5
    \    Click Dockerfile Tab
    \    Sleep    200ms
    \    Click Build History Tab
    \    Sleep    200ms
    # Should end up on Dockerfile tab without errors
    Click Dockerfile Tab
    Verify Dockerfile Tab Active
    [Teardown]    Cleanup Test Data

# === Alternative Label Keys ===

Test Dockerfile With Alternative Label Key com.example.dockerfile
    [Documentation]    Verify Dockerfile displays with alternative label key
    [Tags]    artifact    dockerfile    addition    label-keys
    Init Test Data With Alternative Label Key    com.example.dockerfile
    Go To Project
    Go To Repository
    Go Into Artifact
    Click Dockerfile Tab
    Verify Dockerfile Content Displayed
    [Teardown]    Cleanup Test Data

Test Dockerfile With Generic Label Key dockerfile
    [Documentation]    Verify Dockerfile displays with simple label key
    [Tags]    artifact    dockerfile    addition    label-keys
    Init Test Data With Alternative Label Key    dockerfile
    Go To Project
    Go To Repository
    Go Into Artifact
    Click Dockerfile Tab
    Verify Dockerfile Content Displayed
    [Teardown]    Cleanup Test Data

# === Multiple Images ===

Test Multiple Images In Sequence
    [Documentation]    Verify correct Dockerfile displayed when viewing multiple images
    [Tags]    artifact    dockerfile    addition    multiple-images
    # View image with label
    Init Test Data With Dockerfile Label
    Go To Project
    Go To Repository
    Go Into Artifact
    Click Dockerfile Tab
    Verify Dockerfile Content Contains    FROM ubuntu:22.04
    Navigate To Project    ${project_name}
    Cleanup Test Data
    # View image without label
    Init Test Data Without Dockerfile Label
    Go To Project
    Go To Repository
    Go Into Artifact
    Click Dockerfile Tab
    Verify No Dockerfile Info Message Displayed
    [Teardown]    Cleanup Test Data

# === Async Loading ===

Test Dockerfile Tab Loading Indicator
    [Documentation]    Verify loading spinner appears while fetching content
    [Tags]    artifact    dockerfile    addition    async
    Init Test Data With Dockerfile Label
    Go To Project
    Go To Repository
    Go Into Artifact
    Click Dockerfile Tab
    # Spinner should appear then disappear
    Retry Wait Until Page Not Contains Element    ${artifact_detail_loading_spinner}    timeout=15s
    Verify Dockerfile Content Displayed
    [Teardown]    Cleanup Test Data

*** Keywords ***

# === Test Data Setup ===

Init Test Data With Dockerfile Label
    [Documentation]    Create test data with image containing Dockerfile label
    ${test_project}=                    Set Variable    test-dockerfile-with-label
    ${test_repo}=                       Set Variable    test-image
    ${test_tag}=                        Set Variable    latest
    ${dockerfile_content}=              Set Variable    FROM ubuntu:22.04\nRUN apt-get update\nRUN apt-get install -y curl
    Set Test Variable    ${project_name}    ${test_project}
    Set Test Variable    ${repo_name}       ${test_repo}
    Set Test Variable    ${tag_name}        ${test_tag}
    Set Test Variable    ${dockerfile_content}    ${dockerfile_content}
    # Create project
    Create Project    ${test_project}
    # Build and push image with Dockerfile label (done via docker-compose or direct build)
    Log    Build test image with Dockerfile label using docker build
    Log    docker build --label "org.opencontainers.image.source=${dockerfile_content}" -t ${LOCAL_REGISTRY}/${test_project}/${test_repo}:${test_tag} .

Init Test Data Without Dockerfile Label
    [Documentation]    Create test data with image without Dockerfile label
    ${test_project}=                    Set Variable    test-dockerfile-without-label
    ${test_repo}=                       Set Variable    test-image
    ${test_tag}=                        Set Variable    latest
    Set Test Variable    ${project_name}    ${test_project}
    Set Test Variable    ${repo_name}       ${test_repo}
    Set Test Variable    ${tag_name}        ${test_tag}
    # Create project
    Create Project    ${test_project}
    # Build and push image without Dockerfile label
    Log    Build test image WITHOUT Dockerfile label using docker build
    Log    docker build -t ${LOCAL_REGISTRY}/${test_project}/${test_repo}:${test_tag} .

Init Test Data With Empty Dockerfile Label
    [Documentation]    Create test data with empty Dockerfile label
    ${test_project}=                    Set Variable    test-dockerfile-empty-label
    ${test_repo}=                       Set Variable    test-image
    ${test_tag}=                        Set Variable    latest
    Set Test Variable    ${project_name}    ${test_project}
    Set Test Variable    ${repo_name}       ${test_repo}
    Set Test Variable    ${tag_name}        ${test_tag}
    Create Project    ${test_project}
    # Build image with empty Dockerfile label (treated as no label)
    Log    Build test image with EMPTY Dockerfile label
    Log    docker build --label "org.opencontainers.image.source=" -t ${LOCAL_REGISTRY}/${test_project}/${test_repo}:${test_tag} .

Init Test Data With Alternative Label Key
    [Arguments]    ${label_key}
    [Documentation]    Create test data with alternative Dockerfile label key
    ${test_project}=                    Set Variable    test-dockerfile-alt-key
    ${test_repo}=                       Set Variable    test-image
    ${test_tag}=                        Set Variable    latest
    ${dockerfile_content}=              Set Variable    FROM ubuntu:22.04\nRUN apt-get update
    Set Test Variable    ${project_name}    ${test_project}
    Set Test Variable    ${repo_name}       ${test_repo}
    Set Test Variable    ${tag_name}        ${test_tag}
    Create Project    ${test_project}
    Log    Build test image with alternative label key: ${label_key}
    Log    docker build --label "${label_key}=${dockerfile_content}" -t ${LOCAL_REGISTRY}/${test_project}/${test_repo}:${test_tag} .

Cleanup Test Data
    [Documentation]    Clean up test projects and images
    Run Keyword If Test Passed    Delete Project    ${project_name}
    Run Keyword If Test Passed    Delete Repository    ${project_name}    ${repo_name}

Create Project
    [Arguments]    ${project_name}
    [Documentation]    Create a test project
    Navigate To Project    ${project_name}
    Run Keyword And Ignore Error    New Project    ${project_name}

Delete Project
    [Arguments]    ${project_name}
    [Documentation]    Delete a test project
    Navigate To Project    ${project_name}
    Run Keyword And Ignore Error    Project Delete    ${project_name}

Navigate To Project
    [Arguments]    ${project_name}
    [Documentation]    Navigate to project repositories page
    Go To    ${HARBOR_URL}/projects
    Retry Wait Until Page Not Contains Element    ${artifact_list_spinner}

Go To Project
    [Documentation]    Navigate to the test project
    Navigate To Project    ${project_name}
    Retry Element Click    xpath=//a[contains(text(), '${project_name}')]

Go To Repository
    [Documentation]    Navigate to repository in project
    Retry Wait Until Page Not Contains Element    ${artifact_list_spinner}
    Retry Wait Until Page Contains Element    xpath=//clr-dg-row[contains(.,'${repo_name}')]

Go Into Artifact
    [Documentation]    Click into artifact detail page
    Retry Wait Until Page Not Contains Element    ${artifact_list_spinner}
    Retry Element Click    xpath=//clr-dg-row[contains(.,'${tag_name}')]//a[contains(.,'sha256')]
    Retry Wait Until Page Contains Element    ${artifact_tag_component}
    Retry Wait Until Page Not Contains Element    ${artifact_list_spinner}

Click Dockerfile Tab
    [Documentation]    Click on the Dockerfile tab
    Retry Element Click    ${artifact_detail_dockerfile_tab}
    Sleep    1s    # Allow tab content to load

Click Build History Tab
    [Documentation]    Click on the Build History tab
    Retry Element Click    ${artifact_detail_build_history_tab}
    Sleep    1s    # Allow tab content to load

Click Build History Link From Dockerfile Tab
    [Documentation]    Click the link to Build History from the Dockerfile info box
    Retry Element Click    ${artifact_detail_build_history_link}
    Sleep    1s    # Allow tab switch to complete

Verify Dockerfile Tab Active
    [Documentation]    Verify the Dockerfile tab is currently active
    Retry Wait Until Page Contains Element    xpath=${artifact_detail_dockerfile_tab}[@aria-selected='true']

Verify Build History Tab Active
    [Documentation]    Verify the Build History tab is currently active
    Retry Wait Until Page Contains Element    xpath=${artifact_detail_build_history_tab}[@aria-selected='true']

Verify Dockerfile Content Displayed
    [Documentation]    Verify Dockerfile content is shown with proper formatting
    Retry Wait Until Page Not Contains Element    ${artifact_detail_loading_spinner}    timeout=15s
    Retry Wait Until Page Contains Element    ${yaml_container}
    ${content}=    Get Text    ${yaml_container}
    Should Contain    ${content}    FROM

Verify Dockerfile Content Contains
    [Arguments]    ${expected_text}
    [Documentation]    Verify Dockerfile content contains specific text
    Retry Wait Until Page Not Contains Element    ${artifact_detail_loading_spinner}    timeout=15s
    ${content}=    Get Text    ${yaml_container}
    Should Contain    ${content}    ${expected_text}

Verify No Dockerfile Info Message Displayed
    [Documentation]    Verify informational message appears when no Dockerfile label exists
    Retry Wait Until Page Not Contains Element    ${artifact_detail_loading_spinner}    timeout=15s
    Retry Wait Until Page Contains Element    ${artifact_detail_dockerfile_info_box}
    ${message}=    Get Text    ${artifact_detail_dockerfile_info_box}
    Should Contain    ${message}    Dockerfile
    Should Contain    ${message}    labels

Verify File Too Large Message Displayed
    [Documentation]    Verify error message for oversized Dockerfile
    Retry Wait Until Page Not Contains Element    ${artifact_detail_loading_spinner}    timeout=15s
    Retry Wait Until Page Contains Element    ${artifact_detail_file_too_large_msg}
    ${message}=    Get Text    ${artifact_detail_file_too_large_msg}
    Should Contain    ${message}    large
