// Copyright Project Harbor Authors
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
Library  ../../apitests/python/testutils.py
Library  ../../apitests/python/library/repository.py
Resource  ../../resources/Util.robot
Default Tags  Replication

*** Variables ***
${HARBOR_URL}  https://${ip}
${SSH_USER}  root
${HARBOR_ADMIN}  admin
${SERVER}  ${ip}
${SERVER_URL}  https://${SERVER}
${SERVER_API_ENDPOINT}  ${SERVER_URL}/api
&{SERVER_CONFIG}  endpoint=${SERVER_API_ENDPOINT}  verify_ssl=False
${REMOTE_SERVER}  ${ip1}
${REMOTE_SERVER_URL}  https://${REMOTE_SERVER}
${REMOTE_SERVER_API_ENDPOINT}  ${REMOTE_SERVER_URL}/api

*** Test Cases ***
Test Case - Get Harbor Version
#Just get harbor version and log it
    Get Harbor Version

Test Case - Pro Replication Rules Add
    Init Chrome Driver
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Registries
    Switch To Replication Manage
    Check New Rule UI Without Endpoint
    Close Browser

Test Case - Harbor Endpoint Verification
    #This case need vailid info and selfsign cert
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Registries
    Create A New Endpoint    harbor    edp1${d}    https://${ip}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}    N
    Endpoint Is Pingable
    Enable Certificate Verification
    Endpoint Is Unpingable
    Close Browser

##Test Case - DockerHub Endpoint Add
    #This case need vailid info and selfsign cert
    ##Init Chrome Driver
    ##${d}=    Get Current Date    result_format=%m%s
    ##Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    ##Switch To Registries
    ##Create A New Endpoint    docker-hub    edp1${d}    https://hub.docker.com/    ${DOCKER_USER}    ${DOCKER_PWD}    Y
    ##Close Browser

Test Case - Harbor Endpoint Add
    #This case need vailid info and selfsign cert
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Registries
    Create A New Endpoint    harbor    testabc    https://${ip}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}    Y
    Close Browser

Test Case - Harbor Endpoint Edit
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Registries
    Rename Endpoint  testabc  deletea
    Retry Wait Until Page Contains  deletea
    Close Browser

Test Case - Harbor Endpoint Delete
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Registries
    Delete Endpoint  deletea
    Delete Success  deletea
    Close Browser

Test Case - Replication Rule Edit
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${endpoint1}=    Set Variable    e1${d}
    ${endpoint2}=    Set Variable    e2${d}
    ${rule_name_old}=    Set Variable    rule_testabc${d}
    ${rule_name_new}=    Set Variable    rule_abctest${d}
    ${resource_type}=    Set Variable    chart
    ${dest_namespace}=    Set Variable    dest_namespace${d}
    ${mode}=    Set Variable    Scheduled
    ${cron_str}=    Set Variable    10 10 10 * * *
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Registries
    #Due to docker-hub access limitation, remove docker-hub endpoint
    Create A New Endpoint    harbor    ${endpoint1}    https://cicd.harbor.vmwarecna.net    ${null}    ${null}    Y
    Create A New Endpoint    harbor    ${endpoint2}    https://${ip}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}    Y
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    ${rule_name_old}    pull    nightly/a*    image    ${endpoint1}    project${d}
    Edit Replication Rule  ${rule_name_old}
    #  Change rule-name, source-registry, filter, trigger-mode for edition verification
    Clear Field Of Characters    ${rule_name_input}    30
    Retry Text Input    ${rule_name_input}    ${rule_name_new}
    Select Source Registry  ${endpoint2}
    #Source Resource Filter
    Retry Text Input  ${filter_name_id}  project${d}
    Select From List By Value  ${rule_resource_selector}  ${resource_type}
    Retry Text Input  ${dest_namespace_xpath}  ${dest_namespace}
    Select Trigger  ${mode}
    Retry Text Input  ${targetCron_id}  ${cron_str}
    Retry Double Keywords When Error    Retry Element Click    ${rule_save_button}    Retry Wait Until Page Not Contains Element    ${rule_save_button}
    #  verify all items were changed as expected
    Edit Replication Rule    ${rule_name_new}
    Retry Textfield Value Should Be    ${rule_name_input}               ${rule_name_new}
    Retry List Selection Should Be     ${src_registry_dropdown_list}    ${endpoint2}-https://${ip}
    Retry Textfield Value Should Be    ${filter_name_id}                project${d}
    Retry Textfield Value Should Be    ${dest_namespace_xpath}          ${dest_namespace}
    Retry List Selection Should Be     ${rule_resource_selector}        ${resource_type}
    Retry List Selection Should Be     ${rule_trigger_select}           ${mode}
    Retry Textfield Value Should Be    ${targetCron_id}                 ${cron_str}
    Retry Element Click  ${rule_cancel_btn}
    Delete Replication Rule  ${rule_name_new}
    Close Browser

Test Case - Replication Rule Delete
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${endpoint1}=    Set Variable    e1${d}
    ${rule_name}=    Set Variable    rule_testabc${d}
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Registries
    Create A New Endpoint    harbor    ${endpoint1}    https://${ip}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}    Y
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    ${rule_name}    pull    ${DOCKER_USER}/*    image    ${endpoint1}    project${d}
    Delete Replication Rule  ${rule_name}
    Close Browser

Test Case - Replication Of Pull Images from DockerHub To Self
    @{target_images}=  Create List  mariadb  centos
    &{image1_with_tag}=	 Create Dictionary  image=centos  tag=1.0
    &{image2_with_tag}=	 Create Dictionary  image=mariadb  tag=latest
    ${image1}=  Get From Dictionary  ${image1_with_tag}  image
    ${image1}=  Get Substring  ${image1}  0  -2
    Log All  image1:${image1}
    ${image2}=  Get From Dictionary  ${image2_with_tag}  image
    @{target_images}=  Create List  '&{image1_with_tag}'  '&{image2_with_tag}'
    Body Of Replication Of Pull Images from Registry To Self   docker-hub  https://hub.docker.com/  ${DOCKER_USER}    ${DOCKER_PWD}  ${DOCKER_USER}/{${image1}*,${image2}}  ${null}  N  Flatten 1 Level  @{target_images}

Test Case - Replication Of Push Images from Self To Harbor
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    #login source
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project${d}
    Push Image    ${ip}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}    project${d}    hello-world
    Push Image    ${ip}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}    project${d}    busybox:latest
    Push Image With Tag    ${ip}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}    project${d}    hello-world    v1
    Switch To Registries
    Create A New Endpoint    harbor    e${d}    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    rule${d}    push    project${d}/*    image    e${d}    project_dest${d}
    #logout and login target
    Logout Harbor
    Sign In Harbor    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project_dest${d}
    #logout and login source
    Logout Harbor
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate  rule${d}
    Sleep  20
    Logout Harbor
    Sign In Harbor    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Image Should Be Replicated To Project  project_dest${d}  hello-world
    Image Should Be Replicated To Project  project_dest${d}  busybox
    Close Browser

Test Case - Replication Exclusion Mode And Set Bandwidth
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    # login source
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world:latest
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  busybox:latest
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  alpine:3.10

    # push mode
    Switch To System Labels
    Create New Labels  bad_${d}
    Go Into Project  project${d}
    Go Into Repo  project${d}/busybox
    Add Labels To Tag  latest  bad_${d}
    Switch To Registries
    Create A New Endpoint  harbor  e${d}  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule${d}  push  project${d}/*  image  e${d}  project_dest${d}  filter_tag=3.10  filter_tag_model=excluding  filter_label=bad_${d}  filter_label_model=excluding  bandwidth=100  bandwidth_unit=Kbps
    # logout and login target
    Logout Harbor
    Sign In Harbor    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project_dest${d}
    # logout and login source
    Logout Harbor
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate  rule${d}
    Retry Wait Until Page Contains  Succeeded
    Logout Harbor
    Sign In Harbor    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Image Should Be Replicated To Project  project_dest${d}  hello-world  period=0
    # make sure the excluded image is not replication
    Retry Wait Until Page Contains  1 - 1 of 1 items

    # pull mode
    Create An New Project And Go Into Project  project${d}
    Switch To System Labels
    Create New Labels  bad_${d}
    Switch To Registries
    Create A New Endpoint  harbor  e${d}  https://${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule${d}  pull  project${d}/*  image  e${d}  project${d}  filter_tag=3.10  filter_tag_model=excluding  filter_label=bad_${d}  filter_label_model=excluding  bandwidth=2  bandwidth_unit=Mbps
    Select Rule And Replicate  rule${d}
    Retry Wait Until Page Contains  Succeeded
    Image Should Be Replicated To Project  project${d}  hello-world  period=0
    # make sure the excluded image is not replication
    Retry Wait Until Page Contains  1 - 1 of 1 items
    Close Browser

Test Case - Replication Of Push Chart from Self To Harbor
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    #login source
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project${d}
    Switch To Project Charts
    Upload Chart files
    Switch To Registries
    Create A New Endpoint    harbor    e${d}    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    rule${d}    push    project${d}/*    chart    e${d}    project_dest${d}
    #logout and login target
    Logout Harbor
    Sign In Harbor    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project_dest${d}
    #logout and login source
    Logout Harbor
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate    rule${d}
    Sleep    20
    Logout Harbor
    Sign In Harbor    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Go Into Project    project_dest${d}    has_image=${false}
    Switch To Project Charts
    Go Into Chart Version    ${harbor_chart_name}
    Retry Wait Until Page Contains    ${harbor_chart_version}
    Go Into Chart Detail    ${harbor_chart_version}
    Close Browser

Test Case - Replication Of Push Images from Self To Harbor By Push Event
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=   Set Variable    test_large_image
    ${image_size}=  Set Variable  4096
    ${tag1}=  Set Variable  large_f
    @{tags}   Create List  ${tag1}

    #login source
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project${d}
    Switch To Registries
    Create A New Endpoint    harbor    e${d}    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    rule${d}    push    project${d}/*    image    e${d}    project_dest${d}
    ...    Event Based
    #logout and login target
    Logout Harbor
    Sign In Harbor    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project_dest${d}
    Push Special Image To Project  project${d}  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${image}  tags=@{tags}  size=${image_size}
    # Use tag as identifier for this artifact
    Image Should Be Replicated To Project  project_dest${d}  ${image}  tag=${tag1}  expected_image_size_in_regexp=4(\\\.\\d{1,2})*GiB
    Close Browser

Test Case - Replication Of Pull Images from AWS-ECR To Self
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    #login source
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project${d}
    Switch To Registries
    Create A New Endpoint    aws-ecr    e${d}    us-east-2    ${ecr_ac_id}    ${ecr_ac_key}    Y
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    rule${d}    pull    a/*    image    e${d}    project${d}
    Select Rule And Replicate  rule${d}
    Image Should Be Replicated To Project  project${d}  httpd
    Image Should Be Replicated To Project  project${d}  alpine
    Image Should Be Replicated To Project  project${d}  hello-world
    Close Browser

Test Case - Replication Of Pull Images from Google-GCR To Self
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    #login source
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project    project${d}
    Switch To Registries
    Create A New Endpoint    google-gcr    e${d}    asia.gcr.io    ${null}    ${gcr_ac_key}    Y
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    rule${d}    pull    eminent-nation-87317/*    image    e${d}    project${d}
    Filter Replication Rule  rule${d}
    Select Rule And Replicate  rule${d}
    Image Should Be Replicated To Project  project${d}  httpd
    Image Should Be Replicated To Project  project${d}  tomcat
    Close Browser

Test Case - Replication Of Push Images to DockerHub Triggered By Event
    Body Of Replication Of Push Images to Registry Triggered By Event  docker-hub  https://hub.docker.com/  ${DOCKER_USER}  ${DOCKER_PWD}  ${DOCKER_USER}

#Due to issue of delete event replication
#Test Case - Replication Of Push Images to Google-GCR Triggered By Event
    #Body Of Replication Of Push Images to Registry Triggered By Event  google-gcr  gcr.io  ${null}  ${gcr_ac_key}  eminent-nation-87317/harbor-nightly-replication

Test Case - Replication Of Push Images to AWS-ECR Triggered By Event
    Body Of Replication Of Push Images to Registry Triggered By Event  aws-ecr  us-east-2  ${ecr_ac_id}  ${ecr_ac_key}  harbor-nightly-replication

Test Case - Replication Of Pull Images from Gitlab To Self
    &{image1_with_tag}=	 Create Dictionary  image=photon  tag=1.0
    &{image2_with_tag}=	 Create Dictionary  image=alpine  tag=latest
    ${image1}=  Get From Dictionary  ${image1_with_tag}  image
    ${image2}=  Get From Dictionary  ${image2_with_tag}  image
    @{target_images}=  Create List  '&{image1_with_tag}'  '&{image2_with_tag}'
    Body Of Replication Of Pull Images from Registry To Self   gitlab   https://registry.gitlab.com    ${gitlab_id}    ${gitlab_key}    dannylunsa/test_replication/{${image1},${image2}}  ${null}  N  Flatten All Levels  @{target_images}

Test Case - Replication Of Push Images to Gitlab Triggered By Event
    Body Of Replication Of Push Images to Registry Triggered By Event    gitlab   https://registry.gitlab.com    ${gitlab_id}    ${gitlab_key}    dannylunsa/test_replication

Test Case - Replication Of Pull Manifest List and CNAB from Harbor To Self
    &{image1_with_tag}=	 Create Dictionary  image=busybox  tag=1.32.0  total_artifact_count=9  archive_count=0
    &{image2_with_tag}=	 Create Dictionary  image=index101603308079  tag=index_tag101603308079  total_artifact_count=2  archive_count=0
    &{image3_with_tag}=	 Create Dictionary  image=cnab011609785126  tag=cnab_tag011609785126  total_artifact_count=3  archive_count=2
    ${image1}=  Get From Dictionary  ${image1_with_tag}  image
    ${image2}=  Get From Dictionary  ${image2_with_tag}  image
    ${image3}=  Get From Dictionary  ${image3_with_tag}  image
    @{target_images}=  Create List  '&{image1_with_tag}'  '&{image2_with_tag}'  '&{image3_with_tag}'
    Body Of Replication Of Pull Images from Registry To Self   harbor  https://cicd.harbor.vmwarecna.net  admin  qA5ZgV  nightly/{${image1},${image2},${image3}}  ${null}  Y  Flatten 1 Level  @{target_images}

Test Case - Image Namespace Level Flattening
    [tags]  flattening
    ${src_endpoint}=  Set Variable  ${ip1}

    #Test only for <Flatten All Levels>
    &{image1_with_tag}=	 Create Dictionary  image=test_image_1  tag=tag.1  total_artifact_count=1  archive_count=0
    &{image2_with_tag}=	 Create Dictionary  image=level_1/test_image_2  tag=tag.2  total_artifact_count=1  archive_count=0
    &{image3_with_tag}=	 Create Dictionary  image=level_1/level_2/test_image_3  tag=tag.3  total_artifact_count=1  archive_count=0
    @{src_images}=  Create List  '&{image1_with_tag}'  '&{image2_with_tag}'  '&{image3_with_tag}'
    Replication With Flattening  ${src_endpoint}  10  Flatten All Levels  /  @{src_images}

    #Test only for <Flatten 1 Level>
    &{image1_with_tag}=	 Create Dictionary  image=test_image_1  tag=tag.1  total_artifact_count=1  archive_count=0
    &{image2_with_tag}=	 Create Dictionary  image=level_1/test_image_2  tag=tag.2  total_artifact_count=1  archive_count=0
    &{image3_with_tag}=	 Create Dictionary  image=level_1/level_2/test_image_3  tag=tag.3  total_artifact_count=1  archive_count=0
    @{src_images}=  Create List  '&{image1_with_tag}'  '&{image2_with_tag}'  '&{image3_with_tag}'
    Replication With Flattening  ${src_endpoint}  10  Flatten 1 Level  ${null}  @{src_images}

    #Test only for <Flatten 2 Levels>
    &{image1_with_tag}=	 Create Dictionary  image=level_1/test_image_1  tag=tag.1  total_artifact_count=1  archive_count=0
    &{image2_with_tag}=	 Create Dictionary  image=level_1/level_2/test_image_2  tag=tag.2  total_artifact_count=1  archive_count=0
    &{image3_with_tag}=	 Create Dictionary  image=level_1/level_2/level_3/test_image_3  tag=tag.3  total_artifact_count=1  archive_count=0
    @{src_images}=  Create List  '&{image1_with_tag}'  '&{image2_with_tag}'  '&{image3_with_tag}'
    Replication With Flattening  ${src_endpoint}  10  Flatten 2 Levels  level_1/  @{src_images}

    #Test only for <Flatten 3 Levels>
    &{image1_with_tag}=	 Create Dictionary  image=level_1/level_2/test_image_1  tag=tag.1  total_artifact_count=1  archive_count=0
    &{image2_with_tag}=	 Create Dictionary  image=level_1/level_2/level_3/test_image_2  tag=tag.2  total_artifact_count=1  archive_count=0
    &{image3_with_tag}=	 Create Dictionary  image=level_1/level_2/level_3/level_4/test_image_3  tag=tag.3  total_artifact_count=1  archive_count=0
    @{src_images}=  Create List  '&{image1_with_tag}'  '&{image2_with_tag}'  '&{image3_with_tag}'
    Replication With Flattening  ${src_endpoint}  20  Flatten 3 Levels  level_1/level_2/  @{src_images}

Test Case - Robot Account Do Replication
    [tags]  robot_account_do_replication
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${image1}=  Set Variable  hello-world
    ${tag1}=  Set Variable  latest
    ${image1sha256}=  Set Variable  sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a
    ${image1_short_sha256}=  Get Substring  ${image1sha256}  0  15
    ${image2}=  Set Variable  busybox
    ${tag2}=  Set Variable  latest
    ${image2sha256}=  Set Variable  sha256:34efe68cca33507682b1673c851700ec66839ecf94d19b928176e20d20e02413
    ${image2_short_sha256}=  Get Substring  ${image2sha256}  0  15
    ${index}=  Set Variable  index
    ${index_tag}=  Set Variable  index_tag
    Sign In Harbor    https://${ip1}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project_dest${d}
    # create system Robot Account
    ${robot_account_name}  ${robot_account_secret}=  Create A New System Robot Account  is_cover_all=${true}
    # logout and login source
    Logout Harbor
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    # push mode
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image1}:${tag1}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image2}:${tag2}
    Docker Push Index  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${ip}/project${d}/${index}:${index_tag}  ${ip}/project${d}/${image1}:${tag1}  ${ip}/project${d}/${image2}:${tag2}
    Cosign Generate Key Pair
    Docker Login  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Cosign Sign  ${ip}/project${d}/${image1}:${tag1}
    Cosign Sign  ${ip}/project${d}/${image2}:${tag2}
    Cosign Sign  ${ip}/project${d}/${index}:${index_tag}
    Cosign Sign  ${ip}/project${d}/${index}@${image1sha256}
    Docker Logout  ${ip}
    Switch To Registries
    Create A New Endpoint  harbor  e${d}  https://${ip1}  ${robot_account_name}  ${robot_account_secret}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule_push_${d}  push  project${d}/*  image  e${d}  project_dest${d}
    Select Rule And Replicate  rule_push_${d}
    Retry Wait Until Page Contains  Succeeded
    Logout Harbor
    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Image Should Be Replicated To Project  project_dest${d}  ${image1}  period=0
    Should Be Signed By Cosign  ${tag1}
    Image Should Be Replicated To Project  project_dest${d}  ${image2}  period=0
    Should Be Signed By Cosign  ${tag2}
    Back Project Home  project_dest${d}
    Retry Double Keywords When Error  Go Into Repo  project_dest${d}/${index}  Should Be Signed By Cosign  ${index_tag}
    Back Project Home  project_dest${d}
    Go Into Repo  project_dest${d}/${index}
    Retry Double Keywords When Error  Click Index Achieve  ${index_tag}  Should Be Signed By Cosign  ${image1_short_sha256}
    Back Project Home  project_dest${d}
    Go Into Repo  project_dest${d}/${index}
    Retry Double Keywords When Error  Click Index Achieve  ${index_tag}  Should Not Be Signed By Cosign  ${image2_short_sha256}
    # pull mode
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project_dest${d}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule_pull_${d}  pull  project_dest${d}/*  image  e${d}  project_dest${d}
    Select Rule And Replicate  rule_pull_${d}
    Retry Wait Until Page Contains  Succeeded
    Image Should Be Replicated To Project  project_dest${d}  ${image1}  period=0
    Should Be Signed By Cosign  ${tag1}
    Image Should Be Replicated To Project  project_dest${d}  ${image2}  period=0
    Should Be Signed By Cosign  ${tag2}
    Back Project Home  project_dest${d}
    Retry Double Keywords When Error  Go Into Repo  project_dest${d}/${index}  Should Be Signed By Cosign  ${index_tag}
    Back Project Home  project_dest${d}
    Go Into Repo  project_dest${d}/${index}
    Retry Double Keywords When Error  Click Index Achieve  ${index_tag}  Should Be Signed By Cosign  ${image1_short_sha256}
    Back Project Home  project_dest${d}
    Go Into Repo  project_dest${d}/${index}
    Retry Double Keywords When Error  Click Index Achieve  ${index_tag}  Should Not Be Signed By Cosign  ${image2_short_sha256}
    Close Browser

Test Case - Replication Triggered By Events
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${image1}=  Set Variable  hello-world
    ${tag1}=  Set Variable  latest
    ${image1sha256}=  Set Variable  sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a
    ${image1_short_sha256}=  Get Substring  ${image1sha256}  0  15
    ${image2}=  Set Variable  busybox
    ${tag2}=  Set Variable  latest
    ${image2sha256}=  Set Variable  sha256:34efe68cca33507682b1673c851700ec66839ecf94d19b928176e20d20e02413
    ${image2_short_sha256}=  Get Substring  ${image2sha256}  0  15
    ${index}=  Set Variable  index
    ${index_tag}=  Set Variable  index_tag

    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project_dest${d}
    Logout Harbor

    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Switch To Registries
    Create A New Endpoint  harbor  e${d}  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule_push_${d}  push  project${d}/*  image  e${d}  project_dest${d}  mode=Event Based  del_remote=${true}
    # push
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image1}:${tag1}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image2}:${tag2}    
    Docker Push Index  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${ip}/project${d}/${index}:${index_tag}  ${ip}/project${d}/${image1}:${tag1}  ${ip}/project${d}/${image2}:${tag2}
    Go Into Project  project${d}
    Wait Until Page Contains  project${d}/${image1}
    Wait Until Page Contains  project${d}/${image2}
    Wait Until Page Contains  project${d}/${index}
    Logout Harbor
    
    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Image Should Be Replicated To Project  project_dest${d}  ${image1}  period=0
    Image Should Be Replicated To Project  project_dest${d}  ${image2}  period=0
    Image Should Be Replicated To Project  project_dest${d}  ${index}  period=0
    # sign
    Cosign Generate Key Pair
    Docker Login  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Cosign Sign  ${ip}/project${d}/${image1}:${tag1}
    Cosign Sign  ${ip}/project${d}/${index}:${index_tag}
    Cosign Sign  ${ip}/project${d}/${index}@${image1sha256}
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  project${d}
    Retry Double Keywords When Error  Go Into Repo  project${d}/${image1}  Should Be Signed By Cosign  ${tag1}
    Back Project Home  project${d}
    Retry Double Keywords When Error  Go Into Repo  project${d}/${index}  Should Be Signed By Cosign  ${index_tag}
    Back Project Home  project${d}
    Go Into Repo  project${d}/${index}
    Retry Double Keywords When Error  Click Index Achieve  ${index_tag}  Should Be Signed By Cosign  ${image1_short_sha256}
    Back Project Home  project${d}
    Retry Double Keywords When Error  Go Into Repo  project${d}/${image2}  Should Not Be Signed By Cosign  ${tag2}
    Back Project Home  project${d}
    Go Into Repo  project${d}/${index}
    Retry Double Keywords When Error  Click Index Achieve  ${index_tag}  Should Not Be Signed By Cosign  ${image2_short_sha256}
    Logout Harbor

    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  project_dest${d}
    Retry Double Keywords When Error  Go Into Repo  project_dest${d}/${image1}  Should Be Signed By Cosign  ${tag1}
    Back Project Home  project_dest${d}
    Retry Double Keywords When Error  Go Into Repo  project_dest${d}/${index}  Should Be Signed By Cosign  ${index_tag}
    Back Project Home  project_dest${d}
    Go Into Repo  project_dest${d}/${index}
    Retry Double Keywords When Error  Click Index Achieve  ${index_tag}  Should Be Signed By Cosign  ${image1_short_sha256}
    Back Project Home  project_dest${d}
    Retry Double Keywords When Error  Go Into Repo  project_dest${d}/${image2}  Should Not Be Signed By Cosign  ${tag2}
    Back Project Home  project_dest${d}
    Go Into Repo  project_dest${d}/${index}
    Retry Double Keywords When Error  Click Index Achieve  ${index_tag}  Should Not Be Signed By Cosign  ${image2_short_sha256}
    Logout Harbor
    # delete
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  project${d}
    Delete Repo  project${d}  ${image2}
    Repo Not Exist  project${d}  ${image2}
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image1}
    Retry Double Keywords When Error  Delete Accessory  ${tag1}  Should be Accessory deleted  ${tag1}
    Should Not Be Signed By Cosign  ${tag1}
    Back Project Home  project${d}
    Go Into Repo  project${d}/${index}
    Retry Double Keywords When Error  Delete Accessory  ${index_tag}  Should be Accessory deleted  ${index_tag}
    Should Not Be Signed By Cosign  ${index_tag}
    Click Index Achieve  ${index_tag}
    Retry Double Keywords When Error  Delete Accessory  ${image1_short_sha256}  Should be Accessory deleted  ${image1_short_sha256}
    Should Not Be Signed By Cosign  ${image1_short_sha256}    
    Logout Harbor

    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  project_dest${d}
    Go Into Repo  project_dest${d}/${image2}
    Wait Until Page Contains  We couldn't find any artifacts!
    Back Project Home  project_dest${d}
    Retry Double Keywords When Error  Go Into Repo  project_dest${d}/${image1}  Should be Accessory deleted  ${tag1}    
    Should Not Be Signed By Cosign  ${tag1}
    Back Project Home  project_dest${d}
    Retry Double Keywords When Error  Go Into Repo  project_dest${d}/${index}  Should be Accessory deleted  ${index_tag}
    Should Not Be Signed By Cosign  ${index_tag}
    Retry Double Keywords When Error  Click Index Achieve  ${index_tag}  Should be Accessory deleted  ${image1_short_sha256}
    Should Not Be Signed By Cosign  ${image1_short_sha256}
    Close Browser

Test Case - Enable Replication Of Cosign Deployment Security Policy
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${image1}=  Set Variable  hello-world
    ${tag1}=  Set Variable  latest
    ${image1sha256}=  Set Variable  sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a
    ${image1_short_sha256}=  Get Substring  ${image1sha256}  0  15
    ${image2}=  Set Variable  busybox
    ${tag2}=  Set Variable  latest
    ${image2sha256}=  Set Variable  sha256:34efe68cca33507682b1673c851700ec66839ecf94d19b928176e20d20e02413
    ${image2_short_sha256}=  Get Substring  ${image2sha256}  0  15
    ${index}=  Set Variable  index
    ${index_tag}=  Set Variable  index_tag

    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project_push_dest${d}
    Create An New Project And Go Into Project  project_pull_dest${d}
    Switch To Registries
    Create A New Endpoint  harbor  e${d}  https://${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule_pull_${d}  pull  project${d}/*  image  e${d}  project_pull_dest${d}
    Logout Harbor

    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    # push images
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image1}:${tag1}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image2}:${tag2}
    Docker Push Index  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${ip}/project${d}/${index}:${index_tag}  ${ip}/project${d}/${image1}:${tag1}  ${ip}/project${d}/${image2}:${tag2}
    # enable cosign deployment security policy
    Goto Project Config
    Click Cosign Deployment Security
    Save Project Config
    Content Cosign Deployment security Be Selected
    # push mode replication should fail
    Switch To Registries
    Create A New Endpoint  harbor  e${d}  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule_push_${d}  push  project${d}/*  image  e${d}  project_push_dest${d}
    Select Rule And Replicate  rule_push_${d}
    Check Latest Replication Job Status  Failed
    # pull mode replication should fail
    Logout Harbor
    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate  rule_pull_${d}
    Check Latest Replication Job Status  Failed
    # sign
    Cosign Generate Key Pair
    Docker Login  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Cosign Sign  ${ip}/project${d}/${image1}:${tag1}
    Cosign Sign  ${ip}/project${d}/${image2}:${tag2}
    Cosign Sign  ${ip}/project${d}/${index}:${index_tag}
    Cosign Sign  ${ip}/project${d}/${index}@${image1sha256}
    Cosign Sign  ${ip}/project${d}/${index}@${image2sha256}
    Docker Logout  ${ip}
    # push mode replication should success
    Logout Harbor
    Sign In Harbor  https://${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate  rule_push_${d}
    Check Latest Replication Job Status  Succeeded
    # pull mode replication should success
    Logout Harbor
    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate  rule_pull_${d}
    Check Latest Replication Job Status  Succeeded
    # check project_pull_dest
    Go Into Project  project_pull_dest${d}
    Switch To Project Repo
    Repo Exist  project_pull_dest${d}  ${image1}
    Repo Exist  project_pull_dest${d}  ${image2}
    Repo Exist  project_pull_dest${d}  ${index}
    Retry Double Keywords When Error  Go Into Repo  project_pull_dest${d}/${image1}  Should Be Signed By Cosign  ${tag1}
    Back Project Home  project_pull_dest${d}
    Retry Double Keywords When Error  Go Into Repo  project_pull_dest${d}/${image2}  Should Be Signed By Cosign  ${tag2}
    Back Project Home  project_pull_dest${d}
    Retry Double Keywords When Error  Go Into Repo  project_pull_dest${d}/${index}  Should Be Signed By Cosign  ${index_tag}
    Back Project Home  project_pull_dest${d}
    Go Into Repo  project_pull_dest${d}/${index}
    Retry Double Keywords When Error  Click Index Achieve  ${index_tag}  Should Be Signed By Cosign  ${image1_short_sha256}
    Back Project Home  project_pull_dest${d}
    Go Into Repo  project_pull_dest${d}/${index}
    Retry Double Keywords When Error  Click Index Achieve  ${index_tag}  Should Be Signed By Cosign  ${image2_short_sha256}
    # check project_push_dest
    Go Into Project  project_push_dest${d}
    Switch To Project Repo
    Repo Exist  project_push_dest${d}  ${image1}
    Repo Exist  project_push_dest${d}  ${image2}
    Repo Exist  project_push_dest${d}  ${index}
    Retry Double Keywords When Error  Go Into Repo  project_push_dest${d}/${image1}  Should Be Signed By Cosign  ${tag1}
    Back Project Home  project_push_dest${d}
    Retry Double Keywords When Error  Go Into Repo  project_push_dest${d}/${image2}  Should Be Signed By Cosign  ${tag2}
    Back Project Home  project_push_dest${d}
    Retry Double Keywords When Error  Go Into Repo  project_push_dest${d}/${index}  Should Be Signed By Cosign  ${index_tag}
    Back Project Home  project_push_dest${d}
    Go Into Repo  project_push_dest${d}/${index}
    Retry Double Keywords When Error  Click Index Achieve  ${index_tag}  Should Be Signed By Cosign  ${image1_short_sha256}
    Back Project Home  project_push_dest${d}
    Go Into Repo  project_push_dest${d}/${index}
    Retry Double Keywords When Error  Click Index Achieve  ${index_tag}  Should Be Signed By Cosign  ${image2_short_sha256}
    Close Browser

Test Case - Enable Replication Of Notary Deployment Security Policy
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${image1}=  Set Variable  hello-world
    ${tag1}=  Set Variable  latest
    ${image2}=  Set Variable  busybox
    ${tag2}=  Set Variable  latest

    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project_push_dest${d}
    Create An New Project And Go Into Project  project_pull_dest${d}
    Switch To Registries
    Create A New Endpoint  harbor  e${d}  https://${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule_pull_${d}  pull  project${d}/*  image  e${d}  project_pull_dest${d}
    Logout Harbor

    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    # push images
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image1}:${tag1}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image2}:${tag2}
    # enable notary deployment security policy
    Goto Project Config
    Click Notary Deployment Security
    Save Project Config
    Content Notary Deployment security Be Selected
    # push mode replication should fail
    Switch To Registries
    Create A New Endpoint  harbor  e${d}  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule_push_${d}  push  project${d}/*  image  e${d}  project_push_dest${d}
    Select Rule And Replicate  rule_push_${d}
    Check Latest Replication Job Status  Failed
    # pull mode replication should fail
    Logout Harbor
    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate  rule_pull_${d}
    Check Latest Replication Job Status  Failed
    # sign
    Body Of Admin Push Signed Image  project${d}  ${image1}  ${tag1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Body Of Admin Push Signed Image  project${d}  ${image2}  ${tag2}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    # push mode replication should success
    Logout Harbor
    Sign In Harbor  https://${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate  rule_push_${d}
    Check Latest Replication Job Status  Succeeded
    # pull mode replication should success
    Logout Harbor
    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate  rule_pull_${d}
    Check Latest Replication Job Status  Succeeded
    # check project_pull_dest
    Go Into Project  project_pull_dest${d}
    Switch To Project Repo
    Repo Exist  project_pull_dest${d}  ${image1}
    Repo Exist  project_pull_dest${d}  ${image2}
    # check project_push_dest
    Go Into Project  project_push_dest${d}
    Switch To Project Repo
    Repo Exist  project_push_dest${d}  ${image1}
    Repo Exist  project_push_dest${d}  ${image2}
    Close Browser

Test Case - Enable Replication Of Cosign And Notary Deployment Security Policy
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${image1}=  Set Variable  hello-world
    ${tag1}=  Set Variable  latest
    ${image2}=  Set Variable  busybox
    ${tag2}=  Set Variable  latest

    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project_push_dest${d}
    Create An New Project And Go Into Project  project_pull_dest${d}
    Switch To Registries
    Create A New Endpoint  harbor  e${d}  https://${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule_pull_${d}  pull  project${d}/*  image  e${d}  project_pull_dest${d}
    Logout Harbor

    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    # push images
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image1}:${tag1}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image2}:${tag2}
    # enable cosign deployment security policy
    Goto Project Config
    Click Cosign Deployment Security
    Save Project Config
    Content Cosign Deployment security Be Selected
    # enable notary deployment security policy
    Goto Project Config
    Click Notary Deployment Security
    Save Project Config
    Content Notary Deployment security Be Selected
    # cosign sign
    Cosign Generate Key Pair
    Docker Login  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Cosign Sign  ${ip}/project${d}/${image1}:${tag1}
    Cosign Sign  ${ip}/project${d}/${image2}:${tag2}
    Docker Logout  ${ip}
    # push mode replication should fail
    Switch To Registries
    Create A New Endpoint  harbor  e${d}  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule_push_${d}  push  project${d}/*  image  e${d}  project_push_dest${d}
    Select Rule And Replicate  rule_push_${d}
    Check Latest Replication Job Status  Failed
    # pull mode replication should fail
    Logout Harbor
    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate  rule_pull_${d}
    Check Latest Replication Job Status  Failed
    # notary sign
    Body Of Admin Push Signed Image  project${d}  ${image1}  ${tag1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Body Of Admin Push Signed Image  project${d}  ${image2}  ${tag2}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    # delete cosign accessory
    Logout Harbor
    Sign In Harbor  https://${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image1}
    Retry Double Keywords When Error  Delete Accessory  ${tag1}  Should be Accessory deleted  ${tag1}
    Back Project Home  project${d}
    Go Into Repo  project${d}/${image2}
    Retry Double Keywords When Error  Delete Accessory  ${tag2}  Should be Accessory deleted  ${tag2}
    # push mode replication should fail
    Switch To Replication Manage
    Select Rule And Replicate  rule_push_${d}
    Check Latest Replication Job Status  Failed
    # pull mode replication should fail
    Logout Harbor
    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate  rule_pull_${d}
    Check Latest Replication Job Status  Failed
    # cosign sign
    Docker Login  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Cosign Sign  ${ip}/project${d}/${image1}:${tag1}
    Cosign Sign  ${ip}/project${d}/${image2}:${tag2}
    Docker Logout  ${ip}
    # push mode replication should success
    Logout Harbor
    Sign In Harbor  https://${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate  rule_push_${d}
    Check Latest Replication Job Status  Succeeded
    # pull mode replication should success
    Logout Harbor
    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Select Rule And Replicate  rule_pull_${d}
    Check Latest Replication Job Status  Succeeded
    # check project_pull_dest
    Go Into Project  project_pull_dest${d}
    Switch To Project Repo
    Repo Exist  project_pull_dest${d}  ${image1}
    Repo Exist  project_pull_dest${d}  ${image2}
    Retry Double Keywords When Error  Go Into Repo  project_pull_dest${d}/${image1}  Should Be Signed By Cosign  ${tag1}
    Back Project Home  project_pull_dest${d}
    Retry Double Keywords When Error  Go Into Repo  project_pull_dest${d}/${image2}  Should Be Signed By Cosign  ${tag2}
    # check project_push_dest
    Go Into Project  project_push_dest${d}
    Switch To Project Repo
    Repo Exist  project_push_dest${d}  ${image1}
    Repo Exist  project_push_dest${d}  ${image2}
    Retry Double Keywords When Error  Go Into Repo  project_push_dest${d}/${image1}  Should Be Signed By Cosign  ${tag1}
    Back Project Home  project_push_dest${d}
    Retry Double Keywords When Error  Go Into Repo  project_push_dest${d}/${image2}  Should Be Signed By Cosign  ${tag2}
    Close Browser

Test Case - Carvel Imgpkg Copy To Harbor
    [Tags]  imgpkg_copy
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${out_file_path}=  Set Variable  /tmp/my-bundle.tar
    ${repository}=  Set Variable  my-bundle
    ${tag}=  Set Variable  v1.0.0

    Sign In Harbor  https://${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Logout Harbor
    Sign In Harbor  https://${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project_dest${d}
    Docker Login  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Docker Login  ${ip1}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Prepare Image Package Test Files  ${EXECDIR}/config
    Imgpkg Push  ${ip}  project${d}  ${repository}  ${tag}  ${EXECDIR}/config
    Imgpkg Copy From Registry To Registry  ${ip}/project${d}/${repository}:${tag}  ${ip1}/project_dest${d}/${repository}
    Refresh Repositories
    Repo Exist  project_dest${d}  ${repository}
    Go Into Repo  project_dest${d}/${repository}
    Artifact Exist  ${tag}
    Back Project Home  project_dest${d}
    Delete Repo  project_dest${d}  ${repository}
    Repo Not Exist  project_dest${d}  ${repository}
    Imgpkg Copy From Registry To Local Tarball  ${ip}/project${d}/${repository}:${tag}  ${out_file_path}
    Retry File Should Exist  ${out_file_path}
    Imgpkg Copy From Local Tarball To Registry  ${out_file_path}  ${ip1}/project_dest${d}/${repository}
    Refresh Repositories
    Repo Exist  project_dest${d}  ${repository}
    Retry Element Click  ${repo_search_icon}
    Go Into Repo  project_dest${d}/${repository}
    Artifact Exist  ${tag}
    Docker Logout  ${ip}
    Docker Logout  ${ip1}
    Close Browser