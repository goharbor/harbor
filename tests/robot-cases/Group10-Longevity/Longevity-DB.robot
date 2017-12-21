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
Default Tags  Longevity

Run Regression Test With DB
    [Arguments]  ${ip}
    
    ${HARBOR_URL}=  https://${ip}
    #Test Case - Create An New User
    Init Chrome Driver    
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Close Browser

    #Test Case - Create An New Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Create An New Project  test${d}
    Close Browser	
	
    #Test Case - Push Image
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Create An New Project  test${d}
    Push image  ${ip}  tester${d}  Test1@34  test${d}  hello-world:latest
    Go Into Project  test${d}
    Wait Until Page Contains  test${d}/hello-world

    #Test Case - Manage Project Member
    Init Chrome Driver
    ${d}=    Get current Date  result_format=%m%s
    Create An New Project With New User  url=${HARBOR_URL}  username=alice${d}  email=alice${d}@vmware.com  realname=alice${d}  newPassword=Test1@34  comment=harbor  projectname=project${d}  public=false
    Push image  ip=${ip}  user=alice${d}  pwd=Test1@34  project=project${d}  image=hello-world
    Logout Harbor
    Create An New User  url=${HARBOR_URL}  username=bob${d}  email=bob${d}@vmware.com  realname=bob${d}  newPassword=Test1@34  comment=habor
    Logout Harbor
    Create An New User  url=${HARBOR_URL}  username=carol${d}  email=carol${d}@vmware.com  realname=carol${d}  newPassword=Test1@34  comment=harbor
    Logout Harbor
    User Should Be Owner Of Project  alice${d}  Test1@34  project${d}
    User Should Not Be A Member Of Project  bob${d}  Test1@34  project${d}
    Manage Project Member  alice${d}  Test1@34  project${d}  bob${d}  Add
    User Should Be Guest  bob${d}  Test1@34  project${d}
    Change User Role In Project  alice${d}  Test1@34  project${d}  bob${d}  Developer
    User Should Be Developer  bob${d}  Test1@34  project${d}
    Change User Role In Project  alice${d}  Test1@34  project${d}  bob${d}  Admin
    User Should Be Admin  bob${d}  Test1@34  project${d}  carol${d}
    Manage Project Member  alice${d}  Test1@34  project${d}  bob${d}  Remove
    User Should Not Be A Member Of Project  bob${d}  Test1@34  project${d}
    User Should Be Guest  carol${d}  Test1@34  project${d}
    Close Browser

    #Test Case - Delete A Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New Project With New User  ${HARBOR_URL}  tester${d}  tester${d}@vmware.com  tester${d}  Test1@34  harobr  project${d}  false
    Push Image  ${ip}  tester${d}  Test1@34  project${d}  hello-world  
    Project Should Not Be Deleted  project${d}
    Go Into Project  project${d}
    Delete Repo  project${d}
    Back To projects
    Project Should Be Deleted  project${d}
    Close Browser

    #Test Case - Assign Sys Admin
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Switch to User Tag
    Assign User Admin  tester${d}
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  tester${d}  Test1@34
    Administration Tag Should Display
    Close Browser




