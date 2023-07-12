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
Test Case - Garbage Collection
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    GC Now
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  redis  sha256=e4b315ad03a1d1d9ff0c111e648a1a91066c09ead8352d3d6a48fa971a82922c
    Delete Repo  project${d}  redis
    GC Now
    ${latest_job_id}=  Get Text  ${latest_job_id_xpath}
    Retry GC Should Be Successful  ${latest_job_id}  7 blobs and 1 manifests eligible for deletion
    Retry GC Should Be Successful  ${latest_job_id}  The GC job actual frees up 34 MB space
    Close Browser

Test Case - GC Untagged Images
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    GC Now
    Create An New Project And Go Into Project  project${d}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world  latest
    # make hello-world untagged
    Go Into Repo  project${d}  hello-world
    Go Into Artifact   latest
    Should Contain Tag   latest
    Delete A Tag   latest
    Should Not Contain Tag   latest
    # run gc without param delete untagged artifacts checked,  should not delete hello-world:latest
    GC Now
    ${latest_job_id}=  Get Text  ${latest_job_id_xpath}
    Retry GC Should Be Successful  ${latest_job_id}  ${null}
    Go Into Repo   project${d}  hello-world
    Should Contain Artifact
    # run gc with param delete untagged artifacts checked,  should delete hello-world
    Switch To Garbage Collection
    GC Now  untag=${true}
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
    ${image_a_size}=    Set Variable    321.03MiB
    ${image_a_ver}=  Set Variable  6.8.3
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    GC Now
    Create An New Project And Go Into Project  project${d}  storage_quota=${storage_quota}  storage_quota_unit=${storage_quota_unit}
    Cannot Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image_a}:${image_a_ver}  err_msg=will exceed the configured upper limit of 200.0 MiB
    @{param}  Create List  project${d}
    FOR  ${n}  IN RANGE  1  3
        ${out1}  Run Keyword And Ignore Error  GC Now
        ${latest_job_id}=  Get Text  ${latest_job_id_xpath}
        Retry GC Should Be Successful  ${latest_job_id}  ${null}
        ${out2}  Run Keyword And Ignore Error  Retry Keyword When Return Value Mismatch  Get Project Storage Quota Text From Project Quotas List  0Byte of ${storage_quota}${storage_quota_unit}  2  @{param}
        Exit For Loop If  '${out2[0]}'=='PASS'
        Sleep  1
    END
    Should Be Equal As Strings  '${out2[0]}'  'PASS'
    Close Browser
