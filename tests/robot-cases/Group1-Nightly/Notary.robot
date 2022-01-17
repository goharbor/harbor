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
Test Case - Project Level Policy Notary Deployment security
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world:latest
    Go Into Project  project${d}
    Goto Project Config
    Click Notary Deployment Security
    Save Project Config
    # Verify
    # Unsigned image can not be pulled
    Content Notary Deployment security Be Selected
    Cannot Pull Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world:latest  err_msg=The image is not signed in Notary
    # Signed image can be pulled
    Body Of Admin Push Signed Image  project${d}  redis  latest  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Pull image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  redis  tag=latest
    Close Browser

Test Case - Admin Push Signed Image
    [tags]  sign_image
    Body Of Push Signed Image

Test Case - Admin Push Signed Image And Remove Signature
    [tags]  rm_signature
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    ${user}=  Set Variable  user012
    ${pwd}=   Set Variable  Test1@34
    Sign In Harbor  ${HARBOR_URL}  ${user}  ${pwd}
    Create An New Project And Go Into Project  project${d}
    Body Of Admin Push Signed Image  project${d}  alpine  latest  ${user}  ${pwd}  with_remove=${true}
    Body Of Admin Push Signed Image  project${d}  busybox  latest  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  with_remove=${true}

Test Case - Key Rotate
    [tags]  key_rotate
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    ${user}=  Set Variable  user012
    ${pwd}=   Set Variable  Test1@34
    Sign In Harbor  ${HARBOR_URL}  ${user}  ${pwd}
    Create An New Project And Go Into Project  project${d}
    Body Of Admin Push Signed Image  project${d}  busybox  latest  ${user}  ${pwd}
    Notary Key Rotate   ${ip}  project${d}  busybox  latest  ${user}  ${pwd}
    Body Of Admin Push Signed Image  project${d}  alpine  latest  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Notary Key Rotate   ${ip}  project${d}  alpine  latest  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
