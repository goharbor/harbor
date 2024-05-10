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
${SSH_USER}  root
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Project Quota Sorting
    [Tags]  project_quota_sorting
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    ${d1}=    Get Current Date    result_format=%m%s
    Create An New Project And Go Into Project  project${d1}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d1}  alpine  2.6  2.6
    ${d2}=    Get Current Date    result_format=%m%s
    Create An New Project And Go Into Project  project${d2}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d2}  photon  2.0  2.0
    Switch to Project Quotas Tag
    Check Project Quota Sorting  project${d1}  project${d2}
    Go Into Project  project${d1}
    Delete Repo  project${d1}  alpine
    Go Into Project  project${d2}
    Delete Repo  project${d2}  photon
    GC Now
    Close Browser

Test Case - Garbage Collection
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    GC Now
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  redis  sha256=e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c
    Delete Repo  project${d}  redis
    GC Now  workers=5
    ${latest_job_id}=  Get Text  ${latest_job_id_xpath}
    Retry GC Should Be Successful  ${latest_job_id}  7 blobs and 1 manifests eligible for deletion
    Retry GC Should Be Successful  ${latest_job_id}  The GC job actual frees up 34 MB space
    Close Browser

Test Case - GC Untagged Images
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    GC Now  workers=4
    Create An New Project And Go Into Project  project${d}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world  latest
    # make hello-world untagged
    Go Into Repo  project${d}  hello-world
    Go Into Artifact  latest
    Should Contain Tag  latest
    Delete A Tag  latest
    Should Not Contain Tag  latest
    # run gc without param delete untagged artifacts checked,  should not delete hello-world:latest
    GC Now  workers=3
    ${latest_job_id}=  Get Text  ${latest_job_id_xpath}
    Retry GC Should Be Successful  ${latest_job_id}  ${null}
    Go Into Repo  project${d}  hello-world
    Should Contain Artifact
    # run gc with param delete untagged artifacts checked,  should delete hello-world
    Switch To Garbage Collection
    GC Now  untag=${true}  workers=2
    ${latest_job_id}=  Get Text  ${latest_job_id_xpath}
    Retry GC Should Be Successful  ${latest_job_id}  ${null}
    Go Into Repo  project${d}  hello-world
    Should Not Contain Any Artifact
    Close Browser

# Make sure image logstash was pushed to harbor for the 1st time, so GC will delete it.
Test Case - Project Quotas Control Under GC
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${storage_quota}=  Set Variable  200
    ${storage_quota_unit}=  Set Variable  MiB
    ${image_a}=  Set Variable  logstash
    ${image_a_size}=  Set Variable  321.03MiB
    ${image_a_ver}=  Set Variable  6.8.3
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    GC Now
    Create An New Project And Go Into Project  project${d}  storage_quota=${storage_quota}  storage_quota_unit=${storage_quota_unit}
    Cannot Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image_a}:${image_a_ver}  err_msg=will exceed the configured upper limit of 200.0 MiB
    @{param}  Create List  project${d}
    FOR  ${n}  IN RANGE  1  10
        ${out1}  Run Keyword And Ignore Error  GC Now
        ${latest_job_id}=  Get Text  ${latest_job_id_xpath}
        Retry GC Should Be Successful  ${latest_job_id}  ${null}
        ${out2}  Run Keyword And Ignore Error  Retry Keyword When Return Value Mismatch  Get Project Storage Quota Text From Project Quotas List  0Byte of ${storage_quota}${storage_quota_unit}  2  @{param}
        Exit For Loop If  '${out2[0]}'=='PASS'
        Sleep  5
    END
    Should Be Equal As Strings  '${out2[0]}'  'PASS'
    Close Browser

Test Case - Garbage Collection Accessory
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${image}=  Set Variable  hello-world
    ${tag}=  Set Variable  latest
    ${workers}=  Set Variable  1
    ${deleted_prefix}=  Set Variable  delete blob from storage:
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    ${gc_job_id}=  GC Now
    Wait Until GC Complete  ${gc_job_id}
    Check GC History  ${gc_job_id}  0 blob(s) and 0 manifest(s) deleted, 0 space freed up
    ${log_containing}=  Create List  workers: ${workers}
    ${log_excluding}=  Create List
    Check GC Log  ${gc_job_id}  ${log_containing}  ${log_excluding}

    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}

    ${sbom_digest}  ${signature_digest}  ${signature_of_sbom_digest}  ${signature_of_signature_digest}=  Prepare Accessory  project${d}  ${image}  ${tag}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    # Delete the Signature of Signature
    Delete Accessory By Aeecssory XPath  ${artifact_cosign_cosign_accessory_action_btn}
    ${workers}=  Set Variable  2
    ${gc_job_id}=  GC Now  workers=${workers}
    Wait Until GC Complete  ${gc_job_id}
    Check GC History  ${gc_job_id}  2 blob(s) and 1 manifest(s) deleted
    ${log_containing}=  Create List  ${deleted_prefix} ${signature_of_signature_digest}
    ...  workers: ${workers}
    ${log_excluding}=  Create List  ${deleted_prefix} ${sbom_digest}
    ...  ${deleted_prefix} ${signature_of_sbom_digest}
    ...  ${deleted_prefix} ${signature_digest}
    Check GC Log  ${gc_job_id}  ${log_containing}  ${log_excluding}
    Go Into Repo  project${d}  ${image}
    Retry Button Click  ${artifact_list_accessory_btn}
    # Delete the Signature
    Delete Accessory By Aeecssory XPath  ${artifact_cosign_accessory_action_btn}
    ${workers}=  Set Variable  3
    ${gc_job_id}=  GC Now  workers=${workers}
    Wait Until GC Complete  ${gc_job_id}
    Check GC History  ${gc_job_id}  2 blob(s) and 1 manifest(s) deleted
    ${log_containing}=  Create List  ${deleted_prefix} ${signature_digest}
    ...  workers: ${workers}
    ${log_excluding}=  Create List  ${deleted_prefix} ${sbom_digest}
    ...  ${deleted_prefix} ${signature_of_sbom_digest}
    Check GC Log  ${gc_job_id}  ${log_containing}  ${log_excluding}
    Go Into Repo  project${d}  ${image}
    Retry Button Click  ${artifact_list_accessory_btn}
    # Delete the SBOM
    Delete Accessory By Aeecssory XPath  ${artifact_sbom_accessory_action_btn}
    ${workers}=  Set Variable  4
    ${gc_job_id}=  GC Now  workers=${workers}
    Wait Until GC Complete  ${gc_job_id}
    Check GC History  ${gc_job_id}  4 blob(s) and 2 manifest(s) deleted
    ${log_containing}=  Create List  ${deleted_prefix} ${sbom_digest}
    ...  ${deleted_prefix} ${signature_of_sbom_digest}
    ...  workers: ${workers}
    ${log_excluding}=  Create List
    Check GC Log  ${gc_job_id}  ${log_containing}  ${log_excluding}

    ${sbom_digest}  ${signature_digest}  ${signature_of_sbom_digest}  ${signature_of_signature_digest}=  Prepare Accessory  project${d}  ${image}  ${tag}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    # Delete image tags
    Go Into Repo  project${d}  ${image}
    Go Into Artifact  ${tag}
    Should Contain Tag  ${tag}
    Delete A Tag  ${tag}
    ${workers}=  Set Variable  5
    ${gc_job_id}=  GC Now  workers=${workers}
    Wait Until GC Complete  ${gc_job_id}
    Check GC History  ${gc_job_id}  0 blob(s) and 0 manifest(s) deleted, 0 space freed up
    ${log_containing}=  Create List  workers: ${workers}
    ${log_excluding}=  Create List  ${deleted_prefix} ${sbom_digest}
    ...  ${deleted_prefix} ${signature_digest}
    ...  ${deleted_prefix} ${signature_of_sbom_digest}
    ...  ${deleted_prefix} ${signature_of_signature_digest}
    Check GC Log  ${gc_job_id}  ${log_containing}  ${log_excluding}
    ${workers}=  Set Variable  5
    ${gc_job_id}=  GC Now  workers=${workers}  untag=${true}
    Wait Until GC Complete  ${gc_job_id}
    Check GC History  ${gc_job_id}  10 blob(s) and 5 manifest(s) deleted
    ${log_containing}=  Create List  ${deleted_prefix} ${sbom_digest}
    ...  ${deleted_prefix} ${signature_digest}
    ...  ${deleted_prefix} ${signature_of_sbom_digest}
    ...  ${deleted_prefix} ${signature_of_signature_digest}
    ...  workers: ${workers}
    ${log_excluding}=  Create List
    Check GC Log  ${gc_job_id}  ${log_containing}  ${log_excluding}
    Close Browser
