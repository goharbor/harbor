// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
Resource  ../../resources/Util.robot
Default Tags  BAT

*** Variables ***
${HARBOR_URL}  https://${ip}

*** Test Cases ***
Test Case - Sign With Admin
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Close Browser
	
Test Case - Create An New Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Create An New Project  test${d}
    Close Browser	
	
Test Case - Push Image
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Create An New Project  test${d}

    Push image  ${ip}  tester${d}  Test1@34  test${d}  hello-world:latest
    Go Into Project  test${d}
    Wait Until Page Contains  test${d}/hello-world

Test Case - Project Level Policy Public
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Create An New Project  project${d}
    Go Into Project  project${d}
    Goto Project Config
    Click Project Public
    Save Project Config
    #verify
    Public Should Be Selected 
    Back To Projects
    #project${d}  default should be private
    Project Should Be Public  project${d}
    Close Browser

Test Case - Project Level Policy Content Trust
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Create An New Project  project${d}
    Push Image  ${ip}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}  project${d}  hello-world:latest
    Go Into Project  project${d}
    Goto Project Config
    Click Content Trust
    Save Project Config
    #verify
    Content Trust Should Be Selected
    Cannot Pull Unsigned Image  ${ip}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}  project${d}  hello-world:latest
    Close Browser

Test Case - Create An Replication Rule New Endpoint
    Init Chrome Driver
    ${d}=  Get current date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Create An New Project  project${d}
    Go Into Project  project${d}
    Switch To Replication
    Create An New Rule With New Endpoint  policy_name=test_policy_${d}  policy_description=test_description  destination_name=test_destination_name_${d}  destination_url=test_destination_url_${d}  destination_username=test_destination_username  destination_password=test_destination_password
    Close Browser

Test Case - Scan A Tag
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s
    Create An New Project With New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=tester${d}  newPassword=Test1@34  comment=harbor  projectname=project${d}  public=false
    Push Image  ${ip}  tester${d}  Test1@34  project${d}  hello-world
    Go Into Project  project${d}
    Expand Repo  project${d}
    Scan Repo  latest
    Summary Chart Should Display  latest
    Close Browser

Test Case - Admin Push Signed Image
    Enabe Notary Client

    ${rc}  ${output}=  Run And Return Rc And Output  docker pull hello-world:latest
    Log  ${output}
		
    Push image  ${ip}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}  library  hello-world:latest
    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group9-Content-trust/notary-push-image.sh
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0

    ${rc}  ${output}=  Run And Return Rc And Output  curl -u admin:Harbor12345 -s --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/repositories/library/tomcat/signatures"
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    #Should Contain  ${output}  sha256

Test Case - Admin Push Un-Signed Image	
    ${rc}  ${output}=  Run And Return Rc And Output  docker push ${ip}/library/hello-world:latest
    Log To Console  ${output}