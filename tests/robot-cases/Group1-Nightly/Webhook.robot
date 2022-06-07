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
Test Case - Webhook CRUD
    [Tags]  webhook_crud
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Switch To Project Webhooks
    # create more than one webhooks
    Create A New Webhook  webhook${d}  https://test.com
    Create A New Webhook  webhook2${d}  https://test2.com
    Update A Webhook  webhook${d}  newWebhook${d}  https://new-test.com
    Enable/Deactivate State of Same Webhook  newWebhook${d}
    Delete A Webhook  newWebhook${d}
    Close Browser

Test Case - Artifact Event Type Webhook Functionality
    [Tags]  artifact_webhook
    Init Chrome Driver
    ${image}=  Set Variable  busybox
    ${tag}=  Set Variable  latest
    ${digest}=  Set Variable  sha256:34efe68cca33507682b1673c851700ec66839ecf94d19b928176e20d20e02413
    ${d}=  Get Current Date  result_format=%m%s
    Go To  http://${WEBHOOK_ENDPOINT}
    Delete All Requests
    ${webhook_endpoint_url}=  Get Text  //p//code
    New Tab
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Switch To Project Webhooks
    @{event_type}  Create List  Artifact deleted  Artifact pulled  Artifact pushed
    Create A New Webhook  webhook${d}  ${webhook_endpoint_url}  @{event_type}
    ${handles}=  Get Window Handles
    ${webhook_handle}=  Set Variable  ${handles}[0]
    ${harbor_handle}=  Set Variable  ${handles}[1]
    # Artifact pushed
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}  ${tag}
    Switch Window  ${webhook_handle}
    &{artifact_pushed_property}=  Create Dictionary  type=PUSH_ARTIFACT  operator=${HARBOR_ADMIN}  namespace=project${d}  name=${image}  tag=${tag}
    Verify Request  &{artifact_pushed_property}
    Delete All Requests
    Clean All Local Images
    # Artifact pulled
    Docker Login  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Docker Pull  ${ip}/project${d}/${image}:${tag}
    Docker Logout  ${ip}
    &{artifact_pulled_property}=  Create Dictionary  type=PULL_ARTIFACT  operator=${HARBOR_ADMIN}  namespace=project${d}  name=${image}  tag=${digest}
    Verify Request  &{artifact_pulled_property}
    Delete All Requests
    # Artifact deleted
    Switch Window  ${harbor_handle}
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image}
    @{tag_list}  Create List  ${tag}
    Multi-delete Artifact  @{tag_list}
    Switch Window  ${webhook_handle}
    &{artifact_deleted_property}=  Create Dictionary  type=DELETE_ARTIFACT  operator=${HARBOR_ADMIN}  namespace=project${d}  name=${image}  tag=${tag}
    Verify Request  &{artifact_deleted_property}
    Delete All Requests
    Close Browser

Test Case - Scan Event Type Webhook Functionality
    [Tags]  scan_webhook
    Init Chrome Driver
    ${image1}=  Set Variable  busybox
    ${tag1}=  Set Variable  latest
    ${image2}=  Set Variable  goharbor/harbor-e2e-engine
    ${tag2}=  Set Variable  latest-api
    ${d}=  Get Current Date  result_format=%m%s
    Go To  http://${WEBHOOK_ENDPOINT}
    Delete All Requests
    ${webhook_endpoint_url}=  Get Text  //p//code
    New Tab
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image1}  ${tag1}  ${tag1}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image2}  ${tag2}  ${tag2}
    Switch To Project Webhooks
    @{event_type}  Create List  Scanning finished  Scanning stopped
    Create A New Webhook  webhook${d}  ${webhook_endpoint_url}  @{event_type}
    ${handles}=  Get Window Handles
    ${webhook_handle}=  Set Variable  ${handles}[0]
    ${harbor_handle}=  Set Variable  ${handles}[1]
    # Scanning finished
    Switch To Project Repo
    Go Into Repo  project${d}/${image1}
    Scan Repo  ${tag1}  Succeed
    Switch Window  ${webhook_handle}
    &{scanning_finished_property}=  Create Dictionary  type=SCANNING_COMPLETED  scan_status=Success  namespace=project${d}  tag=${tag1}  name=${image1}
    Verify Request  &{scanning_finished_property}
    Delete All Requests
    # Scanning stopped
    Switch Window  ${harbor_handle}
    Scan Artifact  project${d}  ${image2}
    Stop Scan Artifact
    Check Scan Artifact Job Status Is Stopped
    Switch Window  ${webhook_handle}
    &{scanning_stopped_property}=  Create Dictionary  type=SCANNING_STOPPED  scan_status=Stopped  namespace=project${d}  tag=${tag2}  name=${image2}
    Verify Request  &{scanning_stopped_property}
    Delete All Requests
    Close Browser

Test Case - Tag Retention And Replication Event Type Webhook Functionality
    [Tags]  tag_retention_replication_webhook
    Init Chrome Driver
    ${image}=  Set Variable  busybox
    ${tag1}=  Set Variable  latest
    ${tag2}=  Set Variable  stable
    ${d}=  Get Current Date  result_format=%m%s
    Go To  http://${WEBHOOK_ENDPOINT}
    Delete All Requests
    ${webhook_endpoint_url}=  Get Text  //p//code
    New Tab
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project_dest${d}
    Create An New Project And Go Into Project  project${d}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}  ${tag1}  ${tag1}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}  ${tag2}  ${tag2}
    Switch To Project Webhooks
    @{event_type}  Create List  Tag retention finished  Replication finished
    Create A New Webhook  webhook${d}  ${webhook_endpoint_url}  @{event_type}
    ${handles}=  Get Window Handles
    ${webhook_handle}=  Set Variable  ${handles}[0]
    ${harbor_handle}=  Set Variable  ${handles}[1]
    # Tag retention finished
    Switch To Tag Retention
    Add A Tag Retention Rule
    Edit A Tag Retention Rule  **  ${tag1}
    Execute Run  ${image}
    Switch Window  ${webhook_handle}
    &{tag_retention_finished_property}=  Create Dictionary  type=TAG_RETENTION  operator=MANUAL  project_name=project${d}  name_tag=${image}:${tag2}  status=SUCCESS
    Verify Request  &{tag_retention_finished_property}
    Delete All Requests
    # Replication finished
    Switch Window  ${harbor_handle}
    Switch To Registries
    Create A New Endpoint  harbor  e${d}  https://${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule_push_${d}  push  project${d}/*  image  e${d}  project_push_dest${d}
    Select Rule And Replicate  rule_push_${d}
    Retry Wait Until Page Contains  Succeeded
    Switch Window  ${webhook_handle}
    &{replication_finished_property}=  Create Dictionary  type=REPLICATION  operator=MANUAL  registry_type=harbor  harbor_hostname=${ip}  
    Verify Request  &{replication_finished_property}
    Delete All Requests
    Close Browser

Test Case - Tag Quota Event Type Webhook Functionality
    [Tags]  quota_webhook
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${image}=  Set Variable  nginx
    ${tag1}=  Set Variable  1.17.6
    ${tag2}=  Set Variable  1.14.0
    ${storage_quota}=  Set Variable  50
    Go To  http://${WEBHOOK_ENDPOINT}
    Delete All Requests
    ${webhook_endpoint_url}=  Get Text  //p//code
    New Tab
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}  storage_quota=${storage_quota}  storage_quota_unit=MiB
    Switch To Project Webhooks
    @{event_type}  Create List  Quota near threshold
    Create A New Webhook  webhook${d}  ${webhook_endpoint_url}  @{event_type}
    ${handles}=  Get Window Handles
    ${webhook_handle}=  Set Variable  ${handles}[0]
    ${harbor_handle}=  Set Variable  ${handles}[1]
    # Quota near threshold
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}  ${tag1}  ${tag1}
    Switch Window  ${webhook_handle}
    &{quota_near_threshold_property}=  Create Dictionary  type=QUOTA_WARNING  name=nginx  namespace=project${d}
    Verify Request  &{quota_near_threshold_property}
    Delete All Requests
    # Quota exceed
    Switch Window  ${harbor_handle}
    Delete A Webhook  webhook${d}
    @{event_type}  Create List  Quota exceed
    Create A New Webhook  webhook${d}  ${webhook_endpoint_url}  @{event_type}
    Cannot Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}:${tag2}  err_msg=adding 21.1 MiB of storage resource, which when updated to current usage of 48.5 MiB will exceed the configured upper limit of ${storage_quota}.0 MiB.
    Switch Window  ${webhook_handle}
    &{quota_exceed_property}=  Create Dictionary  type=QUOTA_EXCEED  name=nginx  namespace=project${d}
    Verify Request  &{quota_exceed_property}
    Delete All Requests
    Close Browser