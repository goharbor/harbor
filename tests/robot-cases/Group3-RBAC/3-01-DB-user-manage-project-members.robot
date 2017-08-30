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
Documentation  Manage Project Member
Resource  ../../resources/Util.robot
Default Tags  regression

Test Case - Manage Project Member
    Init Chrome Driver
    ${d}=    Get current Date  result_format=%m%s
    ${rc}  ${ip}=     run and return rc and output  ip add s eth0|grep "inet "|awk '{print $2}'|awk -F "/" '{print $1}'
    log to console  ${ip}
    Create An New User  ${HARBOR_URL}  username=usera${d}  email=usera${d}@vmware.com  realname=usera${d}  newPassword=Test1@34  comment=harbor
    Logout Harbor
    Create An New User  ${HARBOR_URL}  username=userb${d}  email=userb${d}@vmware.com  realname=userb${d}  newPassword=Test1@34  comment=harbor
    Logout Harbor
    Create An New User  ${HARBOR_URL}  username=userc${d}  email=userc${d}@vmware.com  realname=userc${d}  newPassword=Test1@34  comment=harbor
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  usera${d}  Test1@34
    #create project
    Create An New Project  project${d}
    #verify can not change role
    Mouse down  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Mouse up  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Sleep  1
    click element  xpath=//project-detail//li[2]
    page should not contain element  xpath=//project-detail//clr-dg-cell//clr-dg-action-overflow
    Logout Harbor
    #login console as usera and push
    ${rc}=  run and return rc  docker pull hello-world
    ${rc}  ${output}=  run and return rc and output  docker login -u usera${d} -p Test1@34 ${ip}
    ${rc}=  run and return rc  docker tag hello-world ${d}/project${d}/hello-world
    ${rc}=  run and return rc  docker push ${d}/project${d}/hello-world
    ${rc}=  run and return rc  docker logout ${d}
    #logout change userb and pull push
    ${rc}  ${output}=  run and return rc and output docker login -u userb${d} -p Test1@34 ${ip}
    ${rc}=  run and return rc  docker tag hello-world ${d}/project${d}/bbbbb
    ${rc}=  run and return rc  docker pull ${ip}/project${d}/hello-world
    should not be equal as integers  ${rc}  0  
    ${rc}=  run and return rc  docker push ${ip}/project${d}/bbbbb
    should not be equal as integers  ${rc}  0
    #login ui as b
    Sign In Harbor  ${HARBOR_URL}  userb${d}  Test1@34
    page should not contain element  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Logout Harbor
    #login as a
    Sign In Harbor  ${HARBOR_URL}  usera${d}  Test1@34
    Mouse down  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Mouse up  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Sleep  1
    click element  xpath=//project-detail//li[2]
    #click add member
    click element  xpath=//project-detail//button//clr-icon
    Sleep  1
    input text  xpath=//add-member//input[@id="member_name"]  userb${d}
    #select guest
    Mouse down  xpath=//project-detail//form//input[@id="checkrads_guest"]
    Mouse up  xpath=//project-detail//form//input[@id="checkrads_guest"]
    click button  xpath=//project-detail//add-member//button[2]
    Logout Harbor
    #sign in as b
    Sign In Harbor   ${HARBOR_URL}  userb${d}  Test1@34
    #step 12
    page should contain element  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    #step 13
    Mouse down  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Mouse up  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Sleep  1
    click element  xpath=//project-detail//li[2]
    sleep  1
    #page should contain element  xpath=//project-detail//clr-dg-cell//clr-dg-action-overflow[@hidden=""]
    xpath should match x times  //project-detail//clr-dg-action-overflow[@hidden=""]  2
    #step 14
    page should not contain element  xpath=//project-detail//button//clr-icon
    ${rc}  ${output}=  run and return rc and output docker login -u userb${d} -p Test1@34 ${ip}
    #step 15
    ${rc}=  run and return rc  docker pull ${ip}/project${d}/hello-world
    #step 16
    ${rc}=  run and return rc  docker push ${ip}/project${d}/bbbbb
    should not be equal as integers  ${rc}  0
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  usera${d}  Test1@34
    #change userb to developer
    Mouse down  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Mouse up  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Sleep  1
    click element  xpath=//project-detail//li[2]
    sleep  1
    click element  xpath=//project-detail//clr-dg-row-master[contains(.,'userb${d}')]//clr-dg-action-overflow
    click element  xpath=//project-detail//clr-dg-row-master[contains(.,'userb${d}')]//clr-dg-action-overflow//button[contains(.,"Developer")]
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  userb${d}  Test1@34
    page should contain element  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Mouse down  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Mouse up  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Sleep  1
    click element  xpath=//project-detail//li[2]
    sleep  1
    #page should contain element  xpath=//project-detail//clr-dg-cell//clr-dg-action-overflow[@hidden=""]
    xpath should match x times  //project-detail//clr-dg-action-overflow[@hidden=""]  2
    #step 20
    page should not contain element  xpath=//project-detail//button//clr-icon
    #step 21
    ${rc}=  run and return rc  docker login -u userb${d} -p Test1@34 ${ip}
    ${rc}=  run and return rc  docker tag hello-world ${ip}/project${d}/hello-world:v1
    ${rc}=  run and return rc  docker push ${ip}/project${d}/hello-world:v1
    should be equal as integers  ${rc}  0
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  usera${d}  Test1@34
    #step 22
    #change userb to admin of project
    Mouse down  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Mouse up  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Sleep  1
    click element  xpath=//project-detail//li[2]
    sleep  1
    click element  xpath=//project-detail//clr-dg-row-master[contains(.,'userb${d}')]//clr-dg-action-overflow
    click element  xpath=//project-detail//clr-dg-row-master[contains(.,'userb${d}')]//clr-dg-action-overflow//button[contains(.,"Admin")]
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  userb${d}  Test1@34
    page should contain element  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    # add userc
    Mouse down  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Mouse up  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Sleep  1
    click element  xpath=//project-detail//li[2]
    sleep  1
    click element  xpath=//project-detail//button//clr-icon
    input text  xpath=//add-member//input[@id="member_name"]  userc${d}
    mouse down  xpath=//project-detail//form//input[@id="checkrads_guest"]
    mouse up  xpath=//project-detail//form//input[@id="checkrads_guest"]
    click button  xpath=//project-detail//add-member//button[2]
    sleep  1
    #step 25 verify b can change c role
    page should contain element  xpath=//project-detail//clr-dg-row-master[contains(.,'userc${d}')]//clr-dg-action-overflow
    ${rc}=  run and return rc  docker login -u userb${d} -p Test1@34 ${ip}
    ${rc}=  run and return rc  docker tag hello-world ${ip}/project${d}/hello-world:v2
    ${rc}=  run and return rc  docker push ${ip}/project${d}/hello-world:v2
    #should be equal as integers  ${rc}  0
    Logout Harbor
    #step 27 remove b from project
    Sign In Harbor  ${HARBOR_URL}  usera${d}  Test1@34
    Mouse down  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Mouse up  xpath=//clr-dg-cell//a[contains(.,'project${d}')]
    Sleep  1
    click element  xpath=//project-detail//li[2]
    sleep  1
    click element  xpath=//project-detail//clr-dg-row-master[contains(.,'userb${d}')]//clr-dg-action-overflow
    click element  xpath=//project-detail//clr-dg-cell//clr-dg-action-overflow//button[contains(.,"Delete")]   
    sleep  1
    click element  xpath=//confiramtion-dialog//button[2]
    sleep  1
    #step28 
    ${rc}=  run and return rc  docker login -u userb${d} -p Test1@34 ${ip}
    ${rc}=  run and return rc  docker pull ${ip}/project${d}/hello-world
    should not be equal as integers  ${rc}  0
    #step 29
    ${rc}=  run and return rc  docker logout ${ip}
    #step 30
    ${rc}=  run and return rc  docker login -u userc${d} -p Test1@34 ${ip}
    ${rc}=  run and return rc  docker pull ${ip}/project${d}/hello-world
    should be equal as integers  ${rc}  0
    Close Browser