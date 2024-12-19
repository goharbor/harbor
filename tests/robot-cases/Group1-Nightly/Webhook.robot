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
    Create A New Webhook  webhook2${d}  https://test2.com  CloudEvents
    Update A Webhook  webhook${d}  newWebhook${d}  https://new-test.com  CloudEvents
    Enable/Deactivate State of Same Webhook  newWebhook${d}
    Delete A Webhook  newWebhook${d}
    Close Browser

Test Case - Artifact Event Type Webhook Functionality
    [Tags]  artifact_webhook  need_webhook_endpoint
    Init Chrome Driver
    ${image}=  Set Variable  busybox
    ${tag}=  Set Variable  latest
    ${d}=  Get Current Date  result_format=%m%s
    Go To  http://${WEBHOOK_ENDPOINT_UI}
    ${webhook_endpoint_url}=  Get Text  //p//code
    New Tab
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Switch To Project Webhooks
    ${event_type}  Create List  Artifact deleted  Artifact pulled  Artifact pushed
    Create A New Webhook  webhook${d}  ${webhook_endpoint_url}  Default  ${event_type}
    ${handles}=  Get Window Handles
    ${webhook_handle}=  Set Variable  ${handles}[0]
    ${harbor_handle}=  Set Variable  ${handles}[1]
    # Artifact pushed
    Retry Action Keyword  Verify Webhook By Artifact Pushed Event  project${d}  webhook${d}  ${image}  ${tag}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${harbor_handle}  ${webhook_handle}
    # Artifact pulled
    Retry Action Keyword  Verify Webhook By Artifact Pulled Event  project${d}  webhook${d}  ${image}  ${tag}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${harbor_handle}  ${webhook_handle}
    # Artifact deleted
    Retry Action Keyword  Verify Webhook By Artifact Deleted Event  project${d}  webhook${d}  ${image}  ${tag}  ${HARBOR_ADMIN}  ${harbor_handle}  ${webhook_handle}
    Close Browser

Test Case - Scan Event Type Webhook Functionality
    [Tags]  scan_webhook  need_webhook_endpoint
    Init Chrome Driver
    ${image1}=  Set Variable  busybox
    ${tag1}=  Set Variable  latest
    ${image2}=  Set Variable  goharbor/harbor-e2e-engine
    ${tag2}=  Set Variable  latest-api
    ${d}=  Get Current Date  result_format=%m%s
    Go To  http://${WEBHOOK_ENDPOINT_UI}
    ${webhook_endpoint_url}=  Get Text  //p//code
    New Tab
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image1}  ${tag1}  ${tag1}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image2}  ${tag2}  ${tag2}
    Switch To Project Webhooks
    ${event_type}  Create List  Scanning finished  Scanning stopped
    Create A New Webhook  webhook${d}  ${webhook_endpoint_url}  Default  ${event_type}
    ${handles}=  Get Window Handles
    ${webhook_handle}=  Set Variable  ${handles}[0]
    ${harbor_handle}=  Set Variable  ${handles}[1]
    # Scanning finished
    Retry Action Keyword  Verify Webhook By Scanning Finished Event  project${d}  webhook${d}  ${image1}  ${tag1}  ${HARBOR_ADMIN}  ${harbor_handle}  ${webhook_handle}
    # Scanning stopped
    Retry Action Keyword  Verify Webhook By Scanning Stopped Event  project${d}  webhook${d}  ${image2}  ${tag2}  ${HARBOR_ADMIN}  ${harbor_handle}  ${webhook_handle}
    Close Browser

Test Case - Tag Retention And Replication Event Type Webhook Functionality
    [Tags]  tag_retention_replication_webhook  need_webhook_endpoint
    Init Chrome Driver
    ${image}=  Set Variable  busybox
    ${tag1}=  Set Variable  latest
    ${tag2}=  Set Variable  stable
    ${d}=  Get Current Date  result_format=%m%s
    Go To  http://${WEBHOOK_ENDPOINT_UI}
    ${webhook_endpoint_url}=  Get Text  //p//code
    New Tab
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project_dest${d}
    Create An New Project And Go Into Project  project${d}
    Switch To Tag Retention
    Add A Tag Retention Rule
    Edit A Tag Retention Rule  **  ${tag1}
    Switch To Project Webhooks
    ${event_type}  Create List  Tag retention finished  Replication status changed
    Create A New Webhook  webhook${d}  ${webhook_endpoint_url}  Default  ${event_type}
    Switch To Registries
    Create A New Endpoint  harbor  e${d}  https://${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule_push_${d}  push  project${d}/*  image  e${d}  project_push_dest${d}
    ${handles}=  Get Window Handles
    ${webhook_handle}=  Set Variable  ${handles}[0]
    ${harbor_handle}=  Set Variable  ${handles}[1]
    # Tag retention finished
    Retry Action Keyword  Verify Webhook By Tag Retention Finished Event  project${d}  webhook${d}  ${image}  ${tag1}  ${tag2}  ${HARBOR_ADMIN}  ${harbor_handle}  ${webhook_handle}
    # Replication status changed
    Retry Action Keyword  Verify Webhook By Replication Status Changed Event  project${d}  webhook${d}  project_push_dest${d}  rule_push_${d}  ${HARBOR_ADMIN}  ${harbor_handle}  ${webhook_handle}
    Close Browser

Test Case - Tag Quota Event Type Webhook Functionality
    [Tags]  quota_webhook  need_webhook_endpoint
    Init Chrome Driver
    Go To  http://${WEBHOOK_ENDPOINT_UI}
    ${webhook_endpoint_url}=  Get Text  //p//code
    New Tab
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    ${handles}=  Get Window Handles
    ${webhook_handle}=  Set Variable  ${handles}[0]
    ${harbor_handle}=  Set Variable  ${handles}[1]
    Retry Action Keyword  Verify Webhook By Quota Near Threshold Event And Quota Exceed Event  ${webhook_endpoint_url}  ${harbor_handle}  ${webhook_handle}
    Close Browser

Test Case - Artifact Event Type Webhook Functionality By CloudEvents Format
    [Tags]  artifact_webhook_cloudevents  need_webhook_endpoint
    Init Chrome Driver
    ${image}=  Set Variable  busybox
    ${tag}=  Set Variable  latest
    ${payload_format}=  Set Variable  CloudEvents
    ${d}=  Get Current Date  result_format=%m%s
    Go To  http://${WEBHOOK_ENDPOINT_UI}
    ${webhook_endpoint_url}=  Get Text  //p//code
    New Tab
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Switch To Project Webhooks
    ${event_type}  Create List  Artifact deleted  Artifact pulled  Artifact pushed
    Create A New Webhook  webhook${d}  ${webhook_endpoint_url}  ${payload_format}  ${event_type}
    ${handles}=  Get Window Handles
    ${webhook_handle}=  Set Variable  ${handles}[0]
    ${harbor_handle}=  Set Variable  ${handles}[1]
    # Artifact pushed
    Retry Action Keyword  Verify Webhook By Artifact Pushed Event  project${d}  webhook${d}  ${image}  ${tag}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${harbor_handle}  ${webhook_handle}  ${payload_format}
    # Artifact pulled
    Retry Action Keyword  Verify Webhook By Artifact Pulled Event  project${d}  webhook${d}  ${image}  ${tag}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${harbor_handle}  ${webhook_handle}  ${payload_format}
    # Artifact deleted
    Retry Action Keyword  Verify Webhook By Artifact Deleted Event  project${d}  webhook${d}  ${image}  ${tag}  ${HARBOR_ADMIN}  ${harbor_handle}  ${webhook_handle}  ${payload_format}
    Close Browser

Test Case - Scan Event Type Webhook Functionality By CloudEvents Format
    [Tags]  scan_webhook_cloudevents  need_webhook_endpoint
    Init Chrome Driver
    ${image1}=  Set Variable  busybox
    ${tag1}=  Set Variable  latest
    ${image2}=  Set Variable  goharbor/harbor-e2e-engine
    ${tag2}=  Set Variable  5.0.0-api
    ${payload_format}=  Set Variable  CloudEvents
    ${d}=  Get Current Date  result_format=%m%s
    Go To  http://${WEBHOOK_ENDPOINT_UI}
    ${webhook_endpoint_url}=  Get Text  //p//code
    New Tab
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image1}  ${tag1}  ${tag1}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image2}  ${tag2}  ${tag2}
    Switch To Project Webhooks
    ${event_type}  Create List  Scanning finished  Scanning stopped
    Create A New Webhook  webhook${d}  ${webhook_endpoint_url}  ${payload_format}  ${event_type}
    ${handles}=  Get Window Handles
    ${webhook_handle}=  Set Variable  ${handles}[0]
    ${harbor_handle}=  Set Variable  ${handles}[1]
    # Scanning finished
    Retry Action Keyword  Verify Webhook By Scanning Finished Event  project${d}  webhook${d}  ${image1}  ${tag1}  ${HARBOR_ADMIN}  ${harbor_handle}  ${webhook_handle}  ${payload_format}
    # Scanning stopped
    Retry Action Keyword  Verify Webhook By Scanning Stopped Event  project${d}  webhook${d}  ${image2}  ${tag2}  ${HARBOR_ADMIN}  ${harbor_handle}  ${webhook_handle}  ${payload_format}
    Close Browser

Test Case - Tag Retention And Replication Event Type Webhook Functionality By CloudEvents Format
    [Tags]  tag_retention_replication_webhook_cloudevents  need_webhook_endpoint
    Init Chrome Driver
    ${image}=  Set Variable  busybox
    ${tag1}=  Set Variable  latest
    ${tag2}=  Set Variable  stable
    ${payload_format}=  Set Variable  CloudEvents
    ${d}=  Get Current Date  result_format=%m%s
    Go To  http://${WEBHOOK_ENDPOINT_UI}
    ${webhook_endpoint_url}=  Get Text  //p//code
    New Tab
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project_dest${d}
    Create An New Project And Go Into Project  project${d}
    Switch To Tag Retention
    Add A Tag Retention Rule
    Edit A Tag Retention Rule  **  ${tag1}
    Switch To Project Webhooks
    ${event_type}  Create List  Tag retention finished  Replication status changed
    Create A New Webhook  webhook${d}  ${webhook_endpoint_url}  ${payload_format}  ${event_type}
    Switch To Registries
    Create A New Endpoint  harbor  e${d}  https://${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Replication Manage
    Create A Rule With Existing Endpoint  rule_push_${d}  push  project${d}/*  image  e${d}  project_push_dest${d}
    ${handles}=  Get Window Handles
    ${webhook_handle}=  Set Variable  ${handles}[0]
    ${harbor_handle}=  Set Variable  ${handles}[1]
    # Tag retention finished
    Retry Action Keyword  Verify Webhook By Tag Retention Finished Event  project${d}  webhook${d}  ${image}  ${tag1}  ${tag2}  ${HARBOR_ADMIN}  ${harbor_handle}  ${webhook_handle}  ${payload_format}
    # Replication status changed
    Retry Action Keyword  Verify Webhook By Replication Status Changed Event  project${d}  webhook${d}  project_push_dest${d}  rule_push_${d}  ${HARBOR_ADMIN}  ${harbor_handle}  ${webhook_handle}  ${payload_format}
    Close Browser

Test Case - Tag Quota Event Type Webhook Functionality By CloudEvents Format
    [Tags]  quota_webhook_cloudevents  need_webhook_endpoint
    ${payload_format}=  Set Variable  CloudEvents
    Init Chrome Driver
    Go To  http://${WEBHOOK_ENDPOINT_UI}
    ${webhook_endpoint_url}=  Get Text  //p//code
    New Tab
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    ${handles}=  Get Window Handles
    ${webhook_handle}=  Set Variable  ${handles}[0]
    ${harbor_handle}=  Set Variable  ${handles}[1]
    Retry Action Keyword  Verify Webhook By Quota Near Threshold Event And Quota Exceed Event  ${webhook_endpoint_url}  ${harbor_handle}  ${webhook_handle}  ${payload_format}
    Close Browser
