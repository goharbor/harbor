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
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image_argument}
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
    Go Into Project  library
    Go Into Repo  ${image_argument}
    Scan Repo  ${tag}  Succeed
    Move To Summary Chart
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
    Navigate To Projects
    Go Into Project  library
    Go Into Repo  redis
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
    Go Into Project  project${d}
    Go Into Repo  project${d}/tomcat
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
    Navigate To Projects
    Go Into Project  project${d}
    Go Into Repo  memcached
    Scan Result Should Display In List Row  latest
    View Repo Scan Details  @{vulnerability_levels}
    Close Browser

Body Of List Helm Charts
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user027  Test1@34
    Create An New Project And Go Into Project  project${d}

    Switch To Project Charts
    Upload Chart files
    Go Into Chart Version  ${prometheus_chart_name}
    Retry Wait Until Page Contains  ${prometheus_chart_version}
    Go Into Chart Detail  ${prometheus_chart_version}

    # Summary tab
    Retry Wait Until Page Contains Element  ${summary_markdown}
    Retry Wait Until Page Contains Element  ${summary_container}

    # Dependency tab
    Retry Double Keywords When Error  Retry Element Click  xpath=${detail_dependency}  Retry Wait Until Page Contains Element  ${dependency_content}

    # Values tab
    Retry Double Keywords When Error  Retry Element Click  xpath=${detail_value}  Retry Wait Until Page Contains Element  ${value_content}

    Go Into Project  project${d}  has_image=${false}
    Retry Keyword N Times When Error  4  Download Chart File  ${prometheus_chart_name}  ${prometheus_chart_filename}
    Multi-delete Chart Files  ${prometheus_chart_name}  ${harbor_chart_name}
    Close Browser

Body Of Push Signed Image
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    ${user}=  Set Variable  user010
    ${pwd}=   Set Variable  Test1@34
    Sign In Harbor  ${HARBOR_URL}  ${user}  ${pwd}
    Create An New Project And Go Into Project  project${d}
    Body Of Admin Push Signed Image  project${d}  tomcat  latest  ${user}  ${pwd}
    Body Of Admin Push Signed Image  project${d}  alpine  latest  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Close Browser

Body Of Admin Push Signed Image
    [Arguments]  ${project}  ${image}  ${tag}  ${user}  ${pwd}  ${with_remove}=${false}  ${clear_trust_dir}=${true}
    Run Keyword If  ${clear_trust_dir}==${true}  Wait Unitl Command Success  rm -rf ~/.docker/
    ${src_tag}=   Set Variable  latest
    ${src_image}=   Set Variable  ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}:${src_tag}
    Docker Pull  ${src_image}
    Wait Unitl Command Success  ./tests/robot-cases/Group0-Util/notary-push-image.sh ${ip} ${project} ${image} ${tag} ${notaryServerEndpoint} ${src_image} ${user} ${pwd}

    ${rc}  ${output}=  Run And Return Rc And Output  curl -u admin:Harbor12345 -s --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/v2.0/projects/${project}/repositories/${image}/artifacts/${tag}?with_signature=true"

    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  "signed":true

    Run Keyword If  ${with_remove} == ${true}  Notary Remove Signature  ${ip}  ${project}  ${image}  ${tag}  ${user}  ${pwd}

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

Helm CLI Push Without Sign In Harbor
    [Arguments]  ${sign_in_user}  ${sign_in_pwd}
    ${d}=   Get Current Date    result_format=%m%s
    Create An New Project And Go Into Project  project${d}
    Helm Repo Add  ${HARBOR_URL}  ${sign_in_user}  ${sign_in_pwd}  project_name=project${d}
    Helm Repo Push  ${sign_in_user}  ${sign_in_pwd}  ${harbor_chart_filename}
    Switch To Project Charts
    Go Into Chart Version  ${harbor_chart_name}
    Retry Wait Until Page Contains  ${harbor_chart_version}

Helm3 CLI Push Without Sign In Harbor
    [Arguments]  ${sign_in_user}  ${sign_in_pwd}
    ${d}=   Get Current Date    result_format=%m%s
    Create An New Project And Go Into Project  project${d}
    Helm Repo Push  ${sign_in_user}  ${sign_in_pwd}  ${harbor_chart_filename}  helm_repo_name=${HARBOR_URL}/chartrepo/project${d}  helm_cmd=helm3
    Switch To Project Charts
    Retry Double Keywords When Error  Go Into Chart Version  ${harbor_chart_name}  Retry Wait Until Page Contains  ${harbor_chart_version}

Helm3.7 CLI Work Flow
    [Arguments]  ${sign_in_user}  ${sign_in_pwd}
    ${d}=   Get Current Date    result_format=%m%s
    Create An New Project And Go Into Project  project${d}
    Run  rm -rf ./${harbor_helm_name}
    Wait Unitl Command Success  tar zxf ${files_directory}/${harbor_helm_filename}
    Helm3.7 Registry Login  ${ip}  ${sign_in_user}  ${sign_in_pwd}
    Helm3.7 Package  ./${harbor_helm_name}
    Helm3.7 Push  ${harbor_helm_package}  ${ip}  project${d}
    Run  rm -rf ./${harbor_helm_package}
    Retry File Should Not Exist  ./${harbor_helm_package}
    Helm3.7 Pull  ${ip}  project${d}  ${harbor_helm_version}
    Retry File Should Exist  ./${harbor_helm_package}
    Helm3.7 Registry Logout  ${ip}

#Important Note: All CVE IDs in CVE Allowlist cases must unique!
Body Of Verfiy System Level CVE Allowlist
    [Arguments]  ${image_argument}  ${sha256_argument}  ${most_cve_list}  ${single_cve}
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=    Set Variable    ${image_argument}
    ${sha256}=  Set Variable  ${sha256_argument}
    ${signin_user}=    Set Variable  user025
    ${signin_pwd}=    Set Variable  Test1@34
    Sign In Harbor    ${HARBOR_URL}    ${signin_user}    ${signin_pwd}
    Create An New Project And Go Into Project    project${d}
    Push Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    sha256=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  2
    Cannot Pull Image  ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}  err_msg=cannot be pulled due to configured policy
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image}
    Scan Repo  ${sha256}  Succeed
    Logout Harbor

    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Check Listed In CVE Allowlist  project${d}  ${image}  ${sha256}  ${single_cve}  is_in=No
    Switch To Configure
    Switch To Configuration System Setting
    # Add Items To System CVE Allowlist    CVE-2021-36222\nCVE-2021-43527 \nCVE-2021-4044 \nCVE-2021-36084 \nCVE-2021-36085 \nCVE-2021-36086 \nCVE-2021-37750 \nCVE-2021-40528
    Add Items To System CVE Allowlist    ${most_cve_list}
    Cannot Pull Image  ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}  err_msg=cannot be pulled due to configured policy
    # Add Items To System CVE Allowlist    CVE-2021-43519
    Add Items To System CVE Allowlist    ${single_cve}
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Delete Top Item In System CVE Allowlist  count=9
    Cannot Pull Image  ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}  err_msg=cannot be pulled due to configured policy

    Check Listed In CVE Allowlist  project${d}  ${image}  ${sha256}  ${single_cve}
    Close Browser

Body Of Verfiy Project Level CVE Allowlist
    [Arguments]  ${image_argument}  ${sha256_argument}  ${most_cve_list}  ${single_cve}
    [Tags]  run-once
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=    Set Variable    ${image_argument}
    ${sha256}=  Set Variable  ${sha256_argument}
    ${signin_user}=    Set Variable  user025
    ${signin_pwd}=    Set Variable  Test1@34
    Sign In Harbor    ${HARBOR_URL}    ${signin_user}    ${signin_pwd}
    Create An New Project And Go Into Project    project${d}
    Push Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    sha256=${sha256}
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  2
    Cannot Pull Image  ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image}
    Scan Repo  ${sha256}  Succeed
    Go Into Project  project${d}
    Add Items to Project CVE Allowlist    ${most_cve_list}
    Cannot Pull Image  ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Add Items to Project CVE Allowlist    ${single_cve}
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Delete Top Item In Project CVE Allowlist
    Cannot Pull Image  ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Close Browser

Body Of Verfiy Project Level CVE Allowlist By Quick Way of Add System
    [Arguments]  ${image_argument}  ${sha256_argument}  ${cve_list}
    [Tags]  run-once
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=    Set Variable    ${image_argument}
    ${sha256}=  Set Variable  ${sha256_argument}
    ${signin_user}=    Set Variable  user025
    ${signin_pwd}=    Set Variable  Test1@34
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Switch To Configuration System Setting
    Add Items To System CVE Allowlist    ${cve_list}
    Logout Harbor
    Sign In Harbor    ${HARBOR_URL}    ${signin_user}    ${signin_pwd}
    Create An New Project And Go Into Project    project${d}
    Push Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    sha256=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  2
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image}
    Scan Repo  ${sha256}  Succeed
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Go Into Project  project${d}
    Set Project To Project Level CVE Allowlist
    Cannot Pull Image  ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Add System CVE Allowlist to Project CVE Allowlist By Add System Button Click
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
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
        Image Should Be Replicated To Project  ${pro_name}  ${image}  tag=${tag}  total_artifact_count=${total_artifact_count}  archive_count=${archive_count}  times=6
    END

Verify Artifact Display
    [Arguments]  ${pro_name}  @{target_images}
    ${count}=    Get length    ${target_images}
    Should Be True  ${count} > 0
    FOR    ${item}    IN    @{target_images}
        ${item}=  Get Substring  ${item}  1  -1
        ${item}=  Evaluate  ${item}
        ${image}=  Get From Dictionary  ${item}  image
        Image Should Be Replicated To Project  ${pro_name}  ${image}  times=6
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
    Sleep  3
    Switch Window  locator=NEW
    ${Title}=  Get Title
    Should Be Equal  ${Title}  Harbor Swagger
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
    Scan Artifact  project${d}  ${repo}
    Stop Scan Artifact
    Check Scan Artifact Job Status Is Stopped
    # stop scan all
    Scan All Artifact
    Stop Scan All Artifact
    Check Scan All Artifact Job Status Is Stopped
    Close Browser

Prepare Image Package Test Files
    [Arguments]  ${files_path}
    ${rc}  ${output}=  Run And Return Rc And Output  bash tests/robot-cases/Group0-Util/prepare_imgpkg_test_files.sh ${files_path}
