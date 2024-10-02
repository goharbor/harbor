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
Documentation  This resource wrap test case body
Library  ../apitests/python/testutils.py
Library  ../apitests/python/library/repository.py

*** Variables ***

*** Keywords ***
Body Of Manage project publicity
    Init Chrome Driver
    ${d}=    Get Current Date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Create An New Project And Go Into Project  project${d}  public=true

    Push image  ${ip}  user007  Test1@34  project${d}  hello-world:latest
    Pull image  ${ip}  user008  Test1@34  project${d}  hello-world:latest

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Display  project${d}
    Search Private Projects
    Project Should Not Display  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Make Project Private  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Not Display  project${d}
    Cannot Pull Image  ${ip}  user008  Test1@34  project${d}  hello-world:latest  err_msg=unauthorized to access repository

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Make Project Public  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Display  project${d}
    Close Browser

Body Of Scan A Tag In The Repo
    [Arguments]  ${image_argument}  ${tag_argument}  ${is_no_vulerabilty}=${false}
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user023  Test1@34
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  user023  Test1@34  project${d}  ${image_argument}:${tag_argument}
    Go Into Repo  project${d}  ${image_argument}
    Scan Repo  ${tag_argument}  Succeed
    Scan Result Should Display In List Row  ${tag_argument}  is_no_vulerabilty=${is_no_vulerabilty}
    Pull Image  ${ip}  user023  Test1@34  project${d}  ${image_argument}  ${tag_argument}
    # Edit Repo Info
    Close Browser

Body Of Scan Image With Empty Vul
    [Arguments]  ${image_argument}  ${tag_argument}
    Init Chrome Driver
    ${tag}=  Set Variable  ${tag_argument}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  ${image_argument}:${tag_argument}
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Repo  library  ${image_argument}
    Scan Repo  ${tag}  Succeed
    Scan Result Should Display In List Row  ${tag}  is_no_vulerabilty=${true}
    Close Browser

Body Of Manual Scan All
    [Arguments]  @{vulnerability_levels}
    Init Chrome Driver
    ${sha256}=  Set Variable  e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  redis  sha256=${sha256}
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Vulnerability Page
    Trigger Scan Now And Wait Until The Result Appears
    Go Into Repo  library  redis
    Scan Result Should Display In List Row  ${sha256}
    View Repo Scan Details  @{vulnerability_levels}
    Close Browser

Body Of View Scan Results
    [Arguments]  @{vulnerability_levels}
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user025  Test1@34
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  user025  Test1@34  project${d}  tomcat
    Go Into Repo  project${d}  tomcat
    Scan Repo  latest  Succeed
    Scan Result Should Display In List Row  latest
    View Repo Scan Details  @{vulnerability_levels}
    Close Browser

Body Of Scan Image On Push
    [Arguments]  @{vulnerability_levels}
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Goto Project Config
    Enable Scan On Push
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  memcached
    Go Into Repo  project${d}  memcached
    Scan Result Should Display In List Row  latest
    View Repo Scan Details  @{vulnerability_levels}
    Close Browser

Delete A Project Without Sign In Harbor
    [Arguments]  ${harbor_ip}=${ip}  ${username}=${HARBOR_ADMIN}  ${password}=${HARBOR_PASSWORD}
    ${d}=    Get Current Date    result_format=%m%s
    ${project_name}=  Set Variable  000${d}
    ${image}=  Set Variable  hello-world
    Create An New Project And Go Into Project  ${project_name}
    Push Image  ${harbor_ip}  ${username}  ${password}  ${project_name}  ${image}
    Project Should Not Be Deleted  ${project_name}
    Go Into Project  ${project_name}
    Delete Repo  ${project_name}  ${image}
    Navigate To Projects
    Project Should Be Deleted  ${project_name}

Manage Project Member Without Sign In Harbor
    [Arguments]  ${sign_in_user}  ${sign_in_pwd}  ${test_user1}=user005  ${test_user2}=user006  ${is_oidc_mode}=${false}
    ${d}=    Get current Date  result_format=%m%s
    ${image}=  Set Variable  hello-world
    Create An New Project And Go Into Project  project${d}
    Push image  ${ip}  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${image}
    Logout Harbor

    User Should Not Be A Member Of Project  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Manage Project Member  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Add  is_oidc_mode=${is_oidc_mode}
    User Should Be Guest  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Change User Role In Project  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Developer  is_oidc_mode=${is_oidc_mode}
    User Should Be Developer  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Change User Role In Project  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Admin  is_oidc_mode=${is_oidc_mode}
    User Should Be Admin  ${test_user1}  ${sign_in_pwd}  project${d}  ${test_user2}  is_oidc_mode=${is_oidc_mode}
    Change User Role In Project  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Maintainer  is_oidc_mode=${is_oidc_mode}
    User Should Be Maintainer  ${test_user1}  ${sign_in_pwd}  project${d}  ${image}  is_oidc_mode=${is_oidc_mode}
    Manage Project Member  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Remove  is_oidc_mode=${is_oidc_mode}
    User Should Not Be A Member Of Project  ${test_user1}  ${sign_in_pwd}  project${d}    is_oidc_mode=${is_oidc_mode}
    Push image  ${ip}  ${sign_in_user}  ${sign_in_pwd}  project${d}  hello-world
    User Should Be Guest  ${test_user2}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}

Helm CLI Work Flow
    [Arguments]  ${sign_in_user}  ${sign_in_pwd}
    ${d}=   Get Current Date    result_format=%m%s
    Create An New Project And Go Into Project  project${d}
    Run  rm -rf ./${harbor_helm_name}
    Wait Unitl Command Success  tar zxf ${files_directory}/${harbor_helm_filename}
    Helm Registry Login  ${ip}  ${sign_in_user}  ${sign_in_pwd}
    Helm Package  ./${harbor_helm_name}
    Helm Push  ${harbor_helm_package}  ${ip}  project${d}
    Run  rm -rf ./${harbor_helm_package}
    Retry File Should Not Exist  ./${harbor_helm_package}
    Helm Pull  ${ip}  project${d}  ${harbor_helm_version}
    Retry File Should Exist  ./${harbor_helm_package}
    Helm Registry Logout  ${ip}

Body Of Verfiy System Level CVE Allowlist
    [Arguments]  ${image_argument}  ${sha256_argument}  ${most_cve_list}  ${single_cve}
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=    Set Variable    ${image_argument}
    ${sha256}=  Set Variable  ${sha256_argument}
    ${signin_user}=  Set Variable  user025
    ${signin_pwd}=  Set Variable  Test1@34
    Sign In Harbor  ${HARBOR_URL}  ${signin_user}  ${signin_pwd}
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  sha256=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  2
    Cannot Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}  err_msg=cannot be pulled due to configured policy
    Go Into Repo  project${d}  ${image}
    Scan Repo  ${sha256}  Succeed
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Check Listed In CVE Allowlist  project${d}  ${image}  ${sha256}  ${single_cve}  is_in=No
    Switch To Configuration Security
    Retry Wait Element Visible  //li[text()=' None ']
    # Add Items To System CVE Allowlist    CVE-2021-36222\nCVE-2021-43527 \nCVE-2021-4044 \nCVE-2021-36084 \nCVE-2021-36085 \nCVE-2021-36086 \nCVE-2021-37750 \nCVE-2021-40528
    Add Items To System CVE Allowlist  ${most_cve_list}
    Cannot Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}  err_msg=cannot be pulled due to configured policy
    # Add Items To System CVE Allowlist    CVE-2021-43519
    Add Items To System CVE Allowlist  ${single_cve}
    Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}
    # Set System CVE Allowlist expires to expired
    Set CVE Allowlist Expires  ${True}
    Retry Wait Until Page Contains  The system CVE allowlist has expired. You can enable the allowlist by extending the expiration date.
    Cannot Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}  err_msg=cannot be pulled due to configured policy
    # Set System CVE Allowlist expires to not expired
    Set CVE Allowlist Expires  ${False}
    Retry Wait Until Page Does Not Contains  The system CVE allowlist has expired. You can enable the allowlist by extending the expiration date.
    Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}

    Delete Top Item In System CVE Allowlist  count=9
    Cannot Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}  err_msg=cannot be pulled due to configured policy
    Check Listed In CVE Allowlist  project${d}  ${image}  ${sha256}  ${single_cve}
    Close Browser

Body Of Verfiy Project Level CVE Allowlist
    [Arguments]  ${image_argument}  ${sha256_argument}  ${most_cve_list}  ${single_cve}
    [Tags]  run-once
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${image}=  Set Variable  ${image_argument}
    ${sha256}=  Set Variable  ${sha256_argument}
    ${signin_user}=  Set Variable  user025
    ${signin_pwd}=  Set Variable  Test1@34
    Sign In Harbor  ${HARBOR_URL}  ${signin_user}  ${signin_pwd}
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  sha256=${sha256}
    Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  2
    Cannot Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}
    Go Into Repo  project${d}  ${image}
    Scan Repo  ${sha256}  Succeed
    Go Into Project  project${d}
    Add Items to Project CVE Allowlist  ${most_cve_list}
    Cannot Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}
    Add Items to Project CVE Allowlist  ${single_cve}
    Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}
    # Set System CVE Allowlist expires to expired
    Set CVE Allowlist Expires  ${True}
    Retry Wait Until Page Contains  The project CVE allowlist has expired. You can enable the allowlist by extending the expiration date.
    Cannot Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}  err_msg=cannot be pulled due to configured policy
    # Set System CVE Allowlist expires to not expired
    Set CVE Allowlist Expires  ${False}
    Retry Wait Until Page Does Not Contains  The project CVE allowlist has expired. You can enable the allowlist by extending the expiration date.
    Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}
    Delete Top Item In Project CVE Allowlist
    Cannot Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}
    Close Browser

Body Of Verfiy Project Level CVE Allowlist By Quick Way of Add System
    [Arguments]  ${image_argument}  ${sha256_argument}  ${cve_list}
    [Tags]  run-once
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${image}=  Set Variable  ${image_argument}
    ${sha256}=  Set Variable  ${sha256_argument}
    ${signin_user}=  Set Variable  user025
    ${signin_pwd}=   Set Variable  Test1@34
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configuration Security
    Add Items To System CVE Allowlist  ${cve_list}
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${signin_user}  ${signin_pwd}
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  sha256=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  2
    Go Into Repo  project${d}  ${image}
    Scan Repo  ${sha256}  Succeed
    Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}
    Go Into Project  project${d}
    Set Project To Project Level CVE Allowlist
    Cannot Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}
    Add System CVE Allowlist to Project CVE Allowlist By Add System Button Click
    Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}
    # Set System CVE Allowlist expires to expired
    Set CVE Allowlist Expires  ${True}
    Retry Wait Until Page Contains  The project CVE allowlist has expired. You can enable the allowlist by extending the expiration date.
    Cannot Pull Image  ${ip}  ${signin_user}  ${signin_pwd}  project${d}  ${image}  tag=${sha256}  err_msg=cannot be pulled due to configured policy
    # Set System CVE Allowlist expires to not expired
    Set CVE Allowlist Expires  ${False}
    Retry Wait Until Page Does Not Contains  The project CVE allowlist has expired. You can enable the allowlist by extending the expiration date.
    Close Browser

Body Of Replication Of Push Images to Registry Triggered By Event
    [Arguments]  ${provider}  ${endpoint}  ${username}  ${pwd}  ${dest_namespace}  ${image_size}=12
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${sha256}=  Set Variable  0e67625224c1da47cb3270e7a861a83e332f708d3d89dde0cbed432c94824d9a
    ${image}=  Set Variable  test_push_repli
    ${tag1}=  Set Variable  v1.1.0
    @{tags}   Create List  ${tag1}
    #login source
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project${d}
    Switch To Registries
    Create A New Endpoint    ${provider}    e${d}    ${endpoint}    ${username}    ${pwd}    Y
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    rule${d}    push    project${d}/*    image    e${d}    ${dest_namespace}  mode=Event Based  del_remote=${true}
    Push Special Image To Project  project${d}  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${image}  tags=@{tags}  size=${image_size}
    Filter Replication Rule  rule${d}
    Select Rule  rule${d}
    ${endpoint_body}=  Fetch From Right  ${endpoint}  //
    ${dest_namespace}=  Set Variable If  '${provider}'=='gitlab'  ${endpoint_body}/${dest_namespace}  ${dest_namespace}
    Run Keyword If  '${provider}'=='docker-hub' or '${provider}'=='gitlab'  Docker Image Can Be Pulled  ${dest_namespace}/${image}:${tag1}   times=3
    Executions Result Count Should Be  Succeeded  event_based  1
    Go Into Project  project${d}
    Delete Repo  project${d}  ${image}
    Run Keyword If  '${provider}'=='docker-hub' or '${provider}'=='gitlab'  Docker Image Can Not Be Pulled  ${dest_namespace}/${image}:${tag1}
    Switch To Replication Manage
    Filter Replication Rule  rule${d}
    Select Rule  rule${d}
    Executions Result Count Should Be  Succeeded  event_based  2

Body Of Replication Of Pull Images from Registry To Self
    [Arguments]  ${provider}  ${endpoint}  ${username}  ${pwd}  ${src_project_name}  ${des_project_name}  ${verify_verbose}  ${flattening}  @{target_images}
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${_des_pro_name}=  Set Variable If  '${des_project_name}'=='${null}'  project${d}  ${des_project_name}
    #login source
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Run Keyword If  '${des_project_name}'=='${null}'  Create An New Project And Go Into Project  ${_des_pro_name}
    Switch To Registries
    Create A New Endpoint    ${provider}    e${d}    ${endpoint}    ${username}    ${pwd}    Y
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule${d}  pull  ${src_project_name}  all  e${d}  ${_des_pro_name}  flattening=${flattening}
    Select Rule And Replicate  rule${d}
    Check Latest Replication Job Status  Succeeded
    Run Keyword If  '${verify_verbose}'=='Y'  Verify Artifact Display Verbose  ${_des_pro_name}  @{target_images}
    ...  ELSE  Verify Artifact Display  ${_des_pro_name}  @{target_images}
    Close Browser

Verify Artifact Display Verbose
    [Arguments]  ${pro_name}  @{target_images}
    ${count}=    Get length    ${target_images}
    Should Be True  ${count} > 0
    FOR    ${item}    IN    @{target_images}
        ${item}=  Get Substring  ${item}  1  -1
        ${item}=  Evaluate  ${item}
        ${image}=  Get From Dictionary  ${item}  image
        ${tag}=  Get From Dictionary  ${item}  tag
        ${total_artifact_count}=  Get From Dictionary  ${item}  total_artifact_count
        ${archive_count}=  Get From Dictionary  ${item}  archive_count
        Log To Console  Check image ${image}:${tag} replication to Project ${pro_name}
        Image Should Be Replicated To Project  ${pro_name}  ${image}  tag=${tag}  total_artifact_count=${total_artifact_count}  archive_count=${archive_count}
    END

Verify Artifact Display
    [Arguments]  ${pro_name}  @{target_images}
    ${count}=    Get length    ${target_images}
    Should Be True  ${count} > 0
    FOR    ${item}    IN    @{target_images}
        ${item}=  Get Substring  ${item}  1  -1
        ${item}=  Evaluate  ${item}
        ${image}=  Get From Dictionary  ${item}  image
        Image Should Be Replicated To Project  ${pro_name}  ${image}
    END

Replication With Flattening
    [Arguments]  ${src_endpoint}  ${image_size}  ${flattening_type}  ${trimmed_namespace}  @{src_images}
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${src_project}=  Set Variable  project${d}
    Sign In Harbor  https://${src_endpoint}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  ${src_project}
    Close Browser
    FOR    ${item}    IN    @{src_images}
        ${item}=  Get Substring  ${item}  1  -1
        ${item}=  Evaluate  ${item}
        ${image}=  Get From Dictionary  ${item}  image
        ${tag}=  Get From Dictionary  ${item}  tag
        @{tags}   Create List  ${tag}
        Push Special Image To Project  ${src_project}  ${src_endpoint}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${image}  tags=@{tags}  size=${image_size}
    END
    @{target_images}=  Create List
    FOR    ${item}    IN    @{src_images}
        ${item}=  Get Substring  ${item}  1  -1
        ${item}=  Evaluate  ${item}
        ${image}=  Get From Dictionary  ${item}  image
        ${tag}=  Get From Dictionary  ${item}  tag
        ${image}=  Fetch From Right  ${image}  ${trimmed_namespace}
        Log All  ${image}
        &{image_with_tag}=	 Create Dictionary  image=${image}  tag=${tag}
        Append To List  ${target_images}   '&{image_with_tag}'
    END
    Log All  ${target_images}
    Body Of Replication Of Pull Images from Registry To Self   harbor  https://${src_endpoint}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${src_project}/**  ${null}  N  ${flattening_type}  @{target_images}

Check Harbor Api Page
    Retry Link Click  //a[contains(.,'Harbor API V2.0')]
    Switch Window  locator=NEW
    Title Should Be  Harbor Swagger
    Retry Wait Element  xpath=//h2[contains(.,"Harbor API")]

Body Of Stop Scan And Stop Scan All
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s
    ${repo}=    Set Variable    goharbor/harbor-e2e-engine
    ${tag}=    Set Variable    test-ui
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${repo}  ${tag}  ${tag}
    # stop scan
    Retry Action Keyword  Stop Scan  project${d}  ${repo}
    # stop scan all
    Retry Action Keyword  Stop Scan All
    Close Browser

Stop Scan
    [Arguments]  ${project_name}  ${repo}
    Scan Artifact  ${project_name}  ${repo}
    Stop Scan Artifact
    Retry Action Keyword  Check Scan Artifact Job Status Is Stopped

Stop Scan All
    Scan All Artifact
    Stop Scan All Artifact
    Retry Action Keyword  Check Scan All Artifact Job Status Is Stopped

Body Of Generate SBOM of An Image In The Repo
    [Arguments]  ${image_argument}  ${tag_argument}
    Init Chrome Driver

    ${d}=  get current date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image_argument}:${tag_argument}
    Go Into Repo  project${d}  ${image_argument}
    Generate Repo SBOM  ${tag_argument}  Succeed
    Checkout And Review SBOM Details  ${tag_argument}
    Close Browser

Body Of Generate Image SBOM On Push
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Goto Project Config
    Enable Generating SBOM On Push
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  memcached
    Go Into Repo  project${d}  memcached
    Checkout And Review SBOM Details  latest
    Close Browser

Body Of Stop SBOM Manual Generation
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s
    ${repo}=    Set Variable    goharbor/harbor-e2e-engine
    ${tag}=    Set Variable    test-ui
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${repo}  ${tag}  ${tag}
    # stop generate sbom of an artifact
    Retry Action Keyword  Stop SBOM Generation  project${d}  ${repo}
    Close Browser

Stop SBOM Generation
    [Arguments]  ${project_name}  ${repo}
    Generate Artifact SBOM  ${project_name}  ${repo}
    Stop Gen Artifact SBOM
    Retry Action Keyword  Check Gen Artifact SBOM Job Status Is Stopped

Prepare Image Package Test Files
    [Arguments]  ${files_path}
    ${rc}  ${output}=  Run And Return Rc And Output  bash tests/robot-cases/Group0-Util/prepare_imgpkg_test_files.sh ${files_path}

Verify Webhook By Artifact Pushed Event
    [Arguments]  ${project_name}  ${webhook_name}  ${image}  ${tag}  ${user}  ${pwd}  ${harbor_handle}  ${webhook_handle}  ${payload_format}=Default
    &{artifact_pushed_property}=  Create Dictionary
    Run Keyword If  '${payload_format}' == 'Default'  Set To Dictionary  ${artifact_pushed_property}  type=PUSH_ARTIFACT  operator=${user}  namespace=${project_name}  name=${image}  tag=${tag}
    ...  ELSE  Set To Dictionary  ${artifact_pushed_property}  specversion=1.0  type=harbor.artifact.pushed  datacontenttype=application/json  namespace=${project_name}  name=${image}  repo_full_name=${project_name}/${image}  tag=${tag}  operator=${user}
    Switch Window  ${webhook_handle}
    Delete All Requests
    Push Image With Tag  ${ip}  ${user}  ${pwd}  ${project_name}  ${image}  ${tag}
    Switch Window  ${harbor_handle}
    Retry Element Click   xpath=//clr-dg-row[contains(.,'${webhook_name}')]//div[contains(@class,'datagrid-select')]
    ${webhook_execution_id}=  Get Latest Webhook Execution ID
    Retry Action Keyword  Verify Webhook Execution  ${webhook_execution_id}  WEBHOOK  Success  Artifact pushed  ${artifact_pushed_property}
    Verify Webhook Execution Log  ${webhook_execution_id}
    Switch Window  ${webhook_handle}
    Verify Request  &{artifact_pushed_property}
    Clean All Local Images

Verify Webhook By Artifact Pulled Event
    [Arguments]  ${project_name}  ${webhook_name}  ${image}  ${tag}  ${user}  ${pwd}  ${harbor_handle}  ${webhook_handle}  ${payload_format}=Default
    &{artifact_pulled_property}=  Create Dictionary
    Run Keyword If  '${payload_format}' == 'Default'  Set To Dictionary  ${artifact_pulled_property}  type=PULL_ARTIFACT  operator=${user}  namespace=${project_name}  name=${image}
    ...  ELSE  Set To Dictionary  ${artifact_pulled_property}  specversion=1.0  type=harbor.artifact.pulled  datacontenttype=application/json  namespace=${project_name}  name=${image}  repo_full_name=${project_name}/${image}  operator=${user}
    Switch Window  ${webhook_handle}
    Delete All Requests
    Clean All Local Images
    Docker Login  ${ip}  ${user}  ${pwd}
    Docker Pull  ${ip}/${project_name}/${image}:${tag}
    Docker Logout  ${ip}
    Switch Window  ${harbor_handle}
    Go Into Project  ${project_name}
    Switch To Project Webhooks
    Retry Element Click   xpath=//clr-dg-row[contains(.,'${webhook_name}')]//div[contains(@class,'datagrid-select')]
    ${webhook_execution_id}=  Get Latest Webhook Execution ID
    Retry Action Keyword  Verify Webhook Execution  ${webhook_execution_id}  WEBHOOK  Success  Artifact pulled  ${artifact_pulled_property}
    Verify Webhook Execution Log  ${webhook_execution_id}
    Switch Window  ${webhook_handle}
    Verify Request  &{artifact_pulled_property}

Verify Webhook By Artifact Deleted Event
    [Arguments]  ${project_name}  ${webhook_name}  ${image}  ${tag}  ${user}  ${harbor_handle}  ${webhook_handle}  ${payload_format}=Default
    &{artifact_deleted_property}=  Create Dictionary
    Run Keyword If  '${payload_format}' == 'Default'  Set To Dictionary  ${artifact_deleted_property}  type=DELETE_ARTIFACT  operator=${user}  namespace=${project_name}  name=${image}  tag=${tag}
    ...  ELSE  Set To Dictionary  ${artifact_deleted_property}  specversion=1.0  type=harbor.artifact.deleted  datacontenttype=application/json  namespace=${project_name}  name=${image}  repo_full_name=${project_name}/${image}  tag=${tag}  operator=${user}
    Switch Window  ${webhook_handle}
    Delete All Requests
    Switch Window  ${harbor_handle}
    Go Into Repo  ${project_name}  ${image}
    @{tag_list}  Create List  ${tag}
    Multi-delete Artifact  @{tag_list}
    Go Into Project  ${project_name}
    Switch To Project Webhooks
    Retry Element Click   xpath=//clr-dg-row[contains(.,'${webhook_name}')]//div[contains(@class,'datagrid-select')]
    ${webhook_execution_id}=  Get Latest Webhook Execution ID
    Retry Action Keyword  Verify Webhook Execution  ${webhook_execution_id}  WEBHOOK  Success  Artifact deleted  ${artifact_deleted_property}
    Verify Webhook Execution Log  ${webhook_execution_id}
    Switch Window  ${webhook_handle}
    Verify Request  &{artifact_deleted_property}

Verify Webhook By Scanning Finished Event
    [Arguments]  ${project_name}  ${webhook_name}  ${image}  ${tag}  ${user}  ${harbor_handle}  ${webhook_handle}  ${payload_format}=Default
    &{scanning_finished_property}=  Create Dictionary
    Run Keyword If  '${payload_format}' == 'Default'  Set To Dictionary  ${scanning_finished_property}  type=SCANNING_COMPLETED  operator=${user}  scan_status=Success  namespace=${project_name}  tag=${tag}  name=${image}
    ...  ELSE  Set To Dictionary  ${scanning_finished_property}  specversion=1.0  type=harbor.scan.completed  datacontenttype=application/json  operator=${user}  namespace=${project_name}  name=${image}  repo_full_name=${project_name}/${image}  tag=${tag}  scan_status=Success
    Switch Window  ${webhook_handle}
    Delete All Requests
    Switch Window  ${harbor_handle}
    Go Into Repo  ${project_name}  ${image}
    Scan Repo  ${tag}  Succeed
    Go Into Project  ${project_name}
    Switch To Project Webhooks
    Retry Element Click   xpath=//clr-dg-row[contains(.,'${webhook_name}')]//div[contains(@class,'datagrid-select')]
    ${webhook_execution_id}=  Get Latest Webhook Execution ID
    Retry Action Keyword  Verify Webhook Execution  ${webhook_execution_id}  WEBHOOK  Success  Scanning finished  ${scanning_finished_property}
    Verify Webhook Execution Log  ${webhook_execution_id}
    Switch Window  ${webhook_handle}
    Verify Request  &{scanning_finished_property}

Verify Webhook By Scanning Stopped Event
    [Arguments]  ${project_name}  ${webhook_name}  ${image}  ${tag}  ${user}  ${harbor_handle}  ${webhook_handle}  ${payload_format}=Default
    &{scanning_stopped_property}=  Create Dictionary
    Run Keyword If  '${payload_format}' == 'Default'  Set To Dictionary  ${scanning_stopped_property}  type=SCANNING_STOPPED  operator=${user}  scan_status=Stopped  namespace=${project_name}  tag=${tag}  name=${image}
    ...  ELSE  Set To Dictionary  ${scanning_stopped_property}  specversion=1.0  type=harbor.scan.stopped  datacontenttype=application/json  operator=${user}  namespace=${project_name}  name=${image}  repo_full_name=${project_name}/${image}  tag=${tag}  scan_status=Stopped
    Switch Window  ${webhook_handle}
    Delete All Requests
    Switch Window  ${harbor_handle}
    Scan Artifact  ${project_name}  ${image}
    Stop Scan Artifact
    Check Scan Artifact Job Status Is Stopped
    Go Into Project  ${project_name}
    Switch To Project Webhooks
    Retry Element Click   xpath=//clr-dg-row[contains(.,'${webhook_name}')]//div[contains(@class,'datagrid-select')]
    ${webhook_execution_id}=  Get Latest Webhook Execution ID
    Retry Action Keyword  Verify Webhook Execution  ${webhook_execution_id}  WEBHOOK  Success  Scanning stopped  ${scanning_stopped_property}
    Verify Webhook Execution Log  ${webhook_execution_id}
    Switch Window  ${webhook_handle}
    Verify Request  &{scanning_stopped_property}

Verify Webhook By Tag Retention Finished Event
    [Arguments]  ${project_name}  ${webhook_name}  ${image}  ${tag1}  ${tag2}  ${user}  ${harbor_handle}  ${webhook_handle}  ${payload_format}=Default
    &{tag_retention_finished_property}=  Create Dictionary
    Run Keyword If  '${payload_format}' == 'Default'  Set To Dictionary  ${tag_retention_finished_property}  type=TAG_RETENTION  operator=${user}  project_name=${project_name}  name_tag=${image}:${tag2}  status=SUCCESS
    ...  ELSE  Set To Dictionary  ${tag_retention_finished_property}  specversion=1.0  type=harbor.tag_retention.finished  datacontenttype=application/json  operator=${user}  project_name=${project_name}  name_tag=${image}:${tag2}  status=SUCCESS
    Switch Window  ${webhook_handle}
    Delete All Requests
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image}  ${tag1}  ${tag1}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image}  ${tag2}  ${tag2}
    Switch Window  ${harbor_handle}
    Go Into Project  ${project_name}
    Switch To Tag Retention
    Execute Run  ${image}
    Go Into Project  ${project_name}
    Switch To Project Webhooks
    Retry Element Click   xpath=//clr-dg-row[contains(.,'${webhook_name}')]//div[contains(@class,'datagrid-select')]
    ${webhook_execution_id}=  Get Latest Webhook Execution ID
    Retry Action Keyword  Verify Webhook Execution  ${webhook_execution_id}  WEBHOOK  Success  Tag retention finished  ${tag_retention_finished_property}
    Verify Webhook Execution Log  ${webhook_execution_id}
    Switch Window  ${webhook_handle}
    Verify Request  &{tag_retention_finished_property}
    Wait Until Page Contains  "total":2
    Wait Until Page Contains  "retained":1

Verify Webhook By Replication Status Changed Event
    [Arguments]  ${project_name}  ${webhook_name}  ${project_dest_name}  ${replication_rule_name}  ${user}  ${harbor_handle}  ${webhook_handle}  ${payload_format}=Default
    &{replication_finished_property}=  Create Dictionary
    Run Keyword If  '${payload_format}' == 'Default'  Set To Dictionary  ${replication_finished_property}  type=REPLICATION  operator=${user}  registry_type=harbor  harbor_hostname=${ip}
    ...  ELSE  Set To Dictionary  ${replication_finished_property}  specversion=1.0  type=harbor.replication.status.changed  datacontenttype=application/json  operator=${user}  trigger_type=MANUAL  namespace=${project_name}
    Switch Window  ${webhook_handle}
    Delete All Requests
    Switch Window  ${harbor_handle}
    Switch To Replication Manage
    Select Rule And Replicate  ${replication_rule_name}
    Check Latest Replication Job Status  Succeeded
    Go Into Project  ${project_name}
    Switch To Project Webhooks
    Retry Element Click   xpath=//clr-dg-row[contains(.,'${webhook_name}')]//div[contains(@class,'datagrid-select')]
    ${webhook_execution_id}=  Get Latest Webhook Execution ID
    Retry Action Keyword  Verify Webhook Execution  ${webhook_execution_id}  WEBHOOK  Success  Replication status changed  ${replication_finished_property}
    Verify Webhook Execution Log  ${webhook_execution_id}
    Switch Window  ${webhook_handle}
    Verify Request  &{replication_finished_property}

Verify Webhook By Quota Near Threshold Event And Quota Exceed Event
    [Arguments]  ${webhook_endpoint_url}  ${harbor_handle}  ${webhook_handle}  ${payload_format}=Default
    ${d}=  Get Current Date  result_format=%m%s
    ${image}=  Set Variable  nginx
    ${tag1}=  Set Variable  1.17.6
    ${tag2}=  Set Variable  1.14.0
    ${storage_quota}=  Set Variable  50
    Create An New Project And Go Into Project  project${d}  storage_quota=${storage_quota}  storage_quota_unit=MiB
    Switch To Project Webhooks
    ${event_type}  Create List  Quota near threshold
    Create A New Webhook  webhook${d}  ${webhook_endpoint_url}  ${payload_format}  ${event_type}
    &{quota_near_threshold_property}=  Create Dictionary
    Run Keyword If  '${payload_format}' == 'Default'  Set To Dictionary  ${quota_near_threshold_property}  type=QUOTA_WARNING  operator=${HARBOR_ADMIN}  name=nginx  namespace=project${d}
    ...  ELSE  Set To Dictionary  ${quota_near_threshold_property}  specversion=1.0  type=harbor.quota.warned  datacontenttype=application/json  operator=${HARBOR_ADMIN}  name=${image}  repo_full_name=project${d}/${image}  namespace=project${d}
    Switch Window  ${webhook_handle}
    Delete All Requests
    # Quota near threshold
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}  ${tag1}  ${tag1}
    Switch Window  ${harbor_handle}
    Retry Element Click   xpath=//clr-dg-row[contains(.,'webhook${d}')]//div[contains(@class,'datagrid-select')]
    ${webhook_execution_id}=  Get Latest Webhook Execution ID
    Retry Action Keyword  Verify Webhook Execution  ${webhook_execution_id}  WEBHOOK  Success  Quota near threshold  ${quota_near_threshold_property}
    Verify Webhook Execution Log  ${webhook_execution_id}
    Switch Window  ${webhook_handle}
    Verify Request  &{quota_near_threshold_property}
    Retry Action Keyword  Verify Webhook By Quota Exceed Event  project${d}  webhook${d}  ${image}  ${tag2}  ${webhook_endpoint_url}  ${storage_quota}  ${harbor_handle}  ${webhook_handle}  ${payload_format}

Verify Webhook By Quota Exceed Event
    [Arguments]  ${project_name}  ${webhook_name}  ${image}  ${tag}  ${webhook_endpoint_url}  ${storage_quota}  ${harbor_handle}  ${webhook_handle}  ${payload_format}=Default
    &{quota_exceed_property}=  Create Dictionary
    Run Keyword If  '${payload_format}' == 'Default'  Set To Dictionary  ${quota_exceed_property}  type=QUOTA_EXCEED  operator=${HARBOR_ADMIN}  name=${image}  namespace=${project_name}
    ...  ELSE  Set To Dictionary  ${quota_exceed_property}  specversion=1.0  type=harbor.quota.exceeded  datacontenttype=application/json  operator=${HARBOR_ADMIN}  name=${image}  repo_full_name=${project_name}/${image}  namespace=${project_name}
    # Quota exceed
    Switch Window  ${harbor_handle}
    Go Into Project  ${project_name}
    Switch To Project Webhooks
    Delete A Webhook  ${webhook_name}
    ${event_type}  Create List  Quota exceed
    Create A New Webhook  ${webhook_name}  ${webhook_endpoint_url}  ${payload_format}  ${event_type}
    Switch Window  ${webhook_handle}
    Delete All Requests
    Cannot Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${project_name}  ${image}:${tag}
    Switch Window  ${harbor_handle}
    Go Into Project  ${project_name}
    Switch To Project Webhooks
    Retry Element Click   xpath=//clr-dg-row[contains(.,'${webhook_name}')]//div[contains(@class,'datagrid-select')]
    ${webhook_execution_id}=  Get Latest Webhook Execution ID
    Retry Action Keyword  Verify Webhook Execution  ${webhook_execution_id}  WEBHOOK  Success  Quota exceed  ${quota_exceed_property}
    Verify Webhook Execution Log  ${webhook_execution_id}
    Switch Window  ${webhook_handle}
    Verify Request  &{quota_exceed_property}

Create Schedules For Job Service Dashboard Schedules
    [Arguments]  ${project_name}  ${schedule_type}  ${schedule_cron}  ${distribution_endpoint}=${null}
    ${d}=  Get Current Date  result_format=%m%s
    ${distribution_name}=  Set Variable  distribution${d}
    ${distribution_endpoint}=  Set Variable If  "${distribution_endpoint}" == "${null}"  https://${d}  ${distribution_endpoint}
    ${p2p_policy_name}=  Set Variable  policy${d}
    ${replication_policy_name}=  Set Variable  rule${d}
    # Create a retention policy triggered by schedule
    Switch To Tag Retention
    Add A Tag Retention Rule
    Set Tag Retention Policy Schedule  ${schedule_type}  ${schedule_cron}
    # Create a preheat policy triggered by schedule
    Create An New Distribution  Dragonfly  ${distribution_name}  ${distribution_endpoint}
    Go Into Project  ${project_name}
    Create An New P2P Preheat Policy  ${p2p_policy_name}  ${distribution_name}  **  **  Scheduled  ${schedule_type}  ${schedule_cron}
    # Create a replication policy triggered by schedule
    Switch to Registries
    Create A New Endpoint  docker-hub  docker-hub${d}  ${null}  ${null}  ${null}  Y
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  ${replication_policy_name}  pull  goharbor/harbor-core  image  docker-hub${d}  ${project_name}  filter_tag=dev  mode=Scheduled  cron=${schedule_cron}
    # Set up a schedule to scan all
    Switch To Vulnerability Page
    Set Scan Schedule  Custom  value=${schedule_cron}
    # Set up a schedule to GC
    Switch To Garbage Collection
    Set GC Schedule  custom  value=${schedule_cron}
    # Set up a schedule to log rotation
    Switch To Log Rotation
    ${exclude_operations}  Create List  Pull
    Set Log Rotation Schedule  1  Days  ${schedule_type}  ${schedule_cron}  ${exclude_operations}
    [Return]  ${replication_policy_name}  ${p2p_policy_name}  ${distribution_name}

Reset Schedules For Job Service Dashboard Schedules
    [Arguments]  ${project_name}  ${replication_policy_name}  ${p2p_policy_name}
    Go Into Project  ${project_name}
    # Reset the schedule of retention policy
    Switch To Tag Retention
    Set Tag Retention Policy Schedule  None
    # Reset the schedule of preheat policy
    Switch To P2P Preheat
    Delete A P2P Preheat Policy  ${p2p_policy_name}
    # Reset the schedule of replication policy
    Switch To Replication Manage
    Delete Replication Rule  ${replication_policy_name}
    # Reset the schedule of scan all
    Switch To Vulnerability Page
    Set Scan Schedule  None
    # Reset the schedule of GC
    Switch To Garbage Collection
    Set GC Schedule  None
    # Reset the schedule of log rotation
    Switch To Log Rotation
    Set Log Rotation Schedule  2  Days  None

Prepare Accessory
    [Arguments]  ${project_name}  ${image}  ${tag}  ${user}  ${pwd}
    Docker Login  ${ip}  ${user}  ${pwd}
    Cosign Generate Key Pair
    Cosign Sign  ${ip}/${project_name}/${image}:${tag}
    Cosign Push Sbom  ${ip}/${project_name}/${image}:${tag}
    Go Into Repo  ${project_name}  ${image}
    Retry Button Click  ${artifact_list_accessory_btn}
    # Get SBOM digest
    Retry Double Keywords When Error  Retry Button Click  ${artifact_sbom_accessory_action_btn}  Retry Button Click  ${copy_digest_btn}
    Wait Until Element Is Visible And Enabled  ${artifact_digest}
    ${sbom_digest}=  Get Text  ${artifact_digest}
    Retry Double Keywords When Error  Retry Button Click  ${copy_btn}  Retry Wait Element Not Visible  ${copy_btn}
    # Get Signature digest
    Retry Double Keywords When Error  Retry Button Click  ${artifact_cosign_accessory_action_btn}  Retry Button Click  ${copy_digest_btn}
    Wait Until Element Is Visible And Enabled  ${artifact_digest}
    ${signature_digest}=  Get Text  ${artifact_digest}
    Retry Double Keywords When Error  Retry Button Click  ${copy_btn}  Retry Wait Element Not Visible  ${copy_btn}
    Cosign Sign  ${ip}/${project_name}/${image}@${sbom_digest}
    Cosign Sign  ${ip}/${project_name}/${image}@${signature_digest}
    Refresh Artifacts
    Retry Button Click  ${artifact_list_accessory_btn}
    # Get Signature of SBOM digest
    Retry Double Keywords When Error  Retry Button Click  ${artifact_list_sbom_accessory_btn}  Retry Button Click  ${artifact_sbom_cosign_accessory_action_btn}
    Retry Double Keywords When Error  Retry Button Click  ${copy_digest_btn}  Wait Until Element Is Visible And Enabled  ${artifact_digest}
    ${signature_of_sbom_digest}=  Get Text  ${artifact_digest}
    Retry Double Keywords When Error  Retry Button Click  ${copy_btn}  Retry Wait Element Not Visible  ${copy_btn}
    # Get Signature of Signature digest
    Retry Double Keywords When Error  Retry Button Click  ${artifact_list_cosign_accessory_btn}  Retry Button Click  ${artifact_cosign_cosign_accessory_action_btn}
    Retry Double Keywords When Error  Retry Button Click  ${copy_digest_btn}  Wait Until Element Is Visible And Enabled  ${artifact_digest}
    ${signature_of_signature_digest}=  Get Text  ${artifact_digest}
    Retry Double Keywords When Error  Retry Button Click  ${copy_btn}  Retry Wait Element Not Visible  ${copy_btn}
    Docker Logout  ${ip}
    [Return]  ${sbom_digest}  ${signature_digest}  ${signature_of_sbom_digest}  ${signature_of_signature_digest}
