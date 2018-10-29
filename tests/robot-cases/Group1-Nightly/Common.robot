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
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${ip}
${SSH_USER}  root
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Sign With Admin
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Close Browser

Test Case - Vulnerability Data Not Ready
#This case must run before vulnerability db ready
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  library
    Vulnerability Not Ready Project Hint
    Switch To Configure
    Go To Vulnerability Config
    Vulnerability Not Ready Config Hint

Test Case - Garbage Collection
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world
    Sleep  2
    Go Into Project  project${d}
    Delete Repo  project${d}

    Switch To Garbage Collection
    Click GC Now
    Logout Harbor
    Sleep  2
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Garbage Collection
    Sleep  1
    Wait Until Page Contains  Finished

    ${rc}  ${output}=  Run And Return Rc And Output  curl -u ${HARBOR_ADMIN}:${HARBOR_PASSWORD} -i --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/system/gc/1/log"
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  3 blobs eligible for deletion
    Should Contain  ${output}  Deleting blob:
    Should Contain  ${output}  success to run gc in job.

    Close Browser
    
Test Case - Create An New Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  test${d}
    Close Browser

Test Case - Delete A Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world
    Project Should Not Be Deleted  project${d}
    Go Into Project  project${d}
    Delete Repo  project${d}
    Back To projects
    Project Should Be Deleted  project${d}
    Close Browser

Test Case - Read Only Mode
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}

    Enable Read Only
    Cannot Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  busybox:latest

    Disable Read Only
    Sleep  5
    Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  busybox:latest
    Close Browser

Test Case - Repo Size
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  alpine  2.6  2.6
    Go Into Project  library
    Go Into Repo  alpine
    Wait Until Page Contains  1.92MB
    Close Browser

Test Case - Staticsinfo
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Wait Until Element Is Visible  //project/div/div/div[1]/div/statistics-panel/div/div[2]/div[1]/div[2]/div[2]/statistics/div/span[1]
    ${privaterepocount1}=  Get Statics Private Repo
    ${privateprojcount1}=  Get Statics Private Project
    ${publicrepocount1}=  Get Statics Public Repo
    ${publicprojcount1}=  Get Statics Public Project
    ${totalrepocount1}=  Get Statics Total Repo
    ${totalprojcount1}=  Get Statics Total Project
    Create An New Project  private${d}
    Create An New Project  public${d}  true
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  private${d}  hello-world
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  public${d}  hello-world
    Reload Page
    ${privateprojcount}=  evaluate  ${privateprojcount1}+1
    ${privaterepocount}=  evaluate  ${privaterepocount1}+1
    ${publicprojcount}=  evaluate  ${publicprojcount1}+1
    ${publicrepocount}=  evaluate  ${publicrepocount1}+1
    ${totalrepocount}=  evaluate  ${totalrepocount1}+2
    ${totalprojcount}=  evaluate  ${totalprojcount1}+2
    Wait Until Element Is Visible  //project/div/div/div[1]/div/statistics-panel/div/div[2]/div[1]/div[2]/div[2]/statistics/div/span[1]
    ${privaterepocountStr}=  Convert To String  ${privaterepocount}
    Wait Until Element Contains  //project/div/div/div[1]/div/statistics-panel/div/div[2]/div[1]/div[2]/div[2]/statistics/div/span[1]  ${privaterepocountStr}
    ${privaterepocount2}=  Get Statics Private Repo
    ${privateprojcount2}=  get statics private project
    ${publicrepocount2}=  get statics public repo
    ${publicprojcount2}=  get statics public project
    ${totalrepocount2}=  get statics total repo
    ${totalprojcount2}=  get statics total project
    Should Be Equal As Integers  ${privateprojcount2}  ${privateprojcount}
    Should be equal as integers  ${privaterepocount2}  ${privaterepocount}
    Should be equal as integers  ${publicprojcount2}  ${publicprojcount}
    Should be equal as integers  ${publicrepocount2}  ${publicrepocount}
    Should be equal as integers  ${totalprojcount2}  ${totalprojcount}
    Should be equal as integers  ${totalrepocount2}  ${totalrepocount}

Test Case - Push Image
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  test${d}

    Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  test${d}  hello-world:latest
    Go Into Project  test${d}
    Wait Until Page Contains  test${d}/hello-world

Test Case - Project Level Policy Public
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}
    Go Into Project  project${d}
    Goto Project Config
    Click Project Public
    Save Project Config
    # Verify
    Public Should Be Selected
    # Project${d}  default should be private
    # Here logout and login to try avoid a bug only in autotest
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Filter Object  project${d}
    Project Should Be Public  project${d}
    Close Browser

Test Case - Project Level Policy Content Trust
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world:latest
    Go Into Project  project${d}
    Goto Project Config
    Click Content Trust
    Save Project Config
    # Verify
    Content Trust Should Be Selected
    Cannot Pull Unsigned Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world:latest
    Close Browser

Test Case - Verify Download Ca Link
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To System Settings
    Page Should Contain  Registry Root Certificate
    Close Browser

Test Case - Edit Email Settings
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}

    Switch To Email
    Config Email

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}

    Switch To Email
    Verify Email

    Close Browser

Test Case - Edit Token Expire
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To System Settings
    Modify Token Expiration  20
    Logout Harbor

    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To System Settings
    Token Must Be Match  20

    #reset to default
    Modify Token Expiration  30
    Close Browser

Test Case - Create A New Labels
    Init Chrome Driver
    ${d}=    Get Current Date
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To System Labels
    Create New Labels  label_${d}
    Close Browser

Test Case - Update Label
   Init Chrome Driver
   ${d}=    Get Current Date

   Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
   Switch To System Labels
   Create New Labels  label_${d}
   Sleep  3
   ${d1}=    Get Current Date
   Update A Label  label_${d}
   Close Browser

Test Case - Delete Label
    Init Chrome Driver
    ${d}=    Get Current Date
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To System Labels
    Create New Labels  label_${d}
    Sleep  3
    Delete A Label  label_${d}
    Close Browser

Test Case - Disable Scan Schedule
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Go To Vulnerability Config
    Disable Scan Schedule
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Go To Vulnerability Config
    Page Should Contain  None
    Close Browser

Test Case - User View Projects
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user001  Test1@34
    Create An New Project  test${d}1
    Create An New Project  test${d}2
    Create An New Project  test${d}3
    Switch To Log
    Wait Until Page Contains  test${d}1
    Wait Until Page Contains  test${d}2
    Wait Until Page Contains  test${d}3
    Close Browser

Test Case - User View Logs
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user002  Test1@34
    Create An New Project  project${d}
    
    Push image  ${ip}  user002  Test1@34  project${d}  busybox:latest
    Pull image  ${ip}  user002  Test1@34  project${d}  busybox:latest

    Go Into Project  project${d}
    Delete Repo  project${d}

    Sleep  3

    Go To Project Log
    Advanced Search Should Display

    Do Log Advanced Search
    Close Browser


Test Case - Manage Project Member
    Init Chrome Driver
    ${d}=    Get current Date  result_format=%m%s
 
    Sign In Harbor  ${HARBOR_URL}  user004  Test1@34
    Create An New Project  project${d}
    Push image  ip=${ip}  user=user004  pwd=Test1@34  project=project${d}  image=hello-world
    Logout Harbor

    User Should Not Be A Member Of Project  user005  Test1@34  project${d}
    Manage Project Member  user004  Test1@34  project${d}  user005  Add
    User Should Be Guest  user005  Test1@34  project${d}
    Change User Role In Project  user004  Test1@34  project${d}  user005  Developer
    User Should Be Developer  user005  Test1@34  project${d}
    Change User Role In Project  user004  Test1@34  project${d}  user005  Admin
    User Should Be Admin  user005  Test1@34  project${d}  user006
    Manage Project Member  user004  Test1@34  project${d}  user005  Remove
    User Should Not Be A Member Of Project  user005  Test1@34  project${d}
    User Should Be Guest  user006  Test1@34  project${d}

    Close Browser

Test Case - Manage project publicity
    Init Chrome Driver
    ${d}=    Get Current Date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Create An New Project  project${d}  public=true

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
    Cannot Pull image  ${ip}  user008  Test1@34  project${d}  hello-world:latest

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Make Project Public  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Display  project${d}
    Close Browser

Test Case - Assign Sys Admin
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user009  Test1@34
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch to User Tag
    Assign User Admin  user009
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user009  Test1@34
    Administration Tag Should Display
    Close Browser

Test Case - Edit Project Creation
    # Create normal user and login
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user010  Test1@34
    Project Creation Should Display
    Logout Harbor

    Sleep  3
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Set Pro Create Admin Only
    Logout Harbor

    Sign In Harbor  ${HARBOR_URL}  user010  Test1@34
    Project Creation Should Not Display
    Logout Harbor

    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Set Pro Create Every One
    Close browser

Test Case - Edit Repo Info
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    
    Sign In Harbor  ${HARBOR_URL}  user011  Test1@34
    Create An New Project  project${d}
    Push Image  ${ip}  user011  Test1@34  project${d}  hello-world
    Go Into Project  project${d}
    Go Into Repo  project${d}/hello-world
    Edit Repo Info
    Close Browser

Test Case - Delete Multi Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    
    Sign In Harbor  ${HARBOR_URL}  user012  Test1@34
    Create An New Project  projecta${d}
    Create An New Project  projectb${d}
    Push Image  ${ip}  user012  Test1@34  projecta${d}  hello-world
    Filter Object  project
    Wait Until Element Is Not Visible  //clr-datagrid/div/div[2]
    Multi-delete Object  projecta  projectb
    # Verify delete project with image should not be deleted directly
    Delete Fail  projecta${d}
    Delete Success  projectb${d}
    Close Browser

Test Case - Delete Multi Repo
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    
    Sign In Harbor  ${HARBOR_URL}  user013  Test1@34
    Create An New Project  project${d}
    Push Image  ${ip}  user013  Test1@34  project${d}  hello-world
    Push Image  ${ip}  user013  Test1@34  project${d}  busybox
    Sleep  2
    Go Into Project  project${d}
    Multi-delete Object  hello-world  busybox
    # Verify
    Delete Success  hello-world  busybox
    Close Browser

Test Case - Delete Multi Tag
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    
    Sign In Harbor  ${HARBOR_URL}  user014  Test1@34
    Create An New Project  project${d}
    Push Image With Tag  ${ip}  user014  Test1@34  project${d}  redis  3.2.10-alpine  3.2.10-alpine
    Push Image With Tag  ${ip}  user014  Test1@34  project${d}  redis  4.0.7-alpine  4.0.7-alpine
    Sleep  2
    Go Into Project  project${d}
    Go Into Repo  redis
    Multi-delete object  3.2.10-alpine  4.0.7-alpine
    # Verify
    Delete Success  3.2.10-alpine  4.0.7-alpine
    Close Browser

Test Case - Delete Repo on CardView
    Init Chrome Driver
    ${d}=   Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user015  Test1@34
    Create An New Project  project${d}
    Push Image  ${ip}  user015  Test1@34  project${d}  hello-world
    Push Image  ${ip}  user015  Test1@34  project${d}  busybox
    Sleep  2
    Go Into Project  project${d}
    Switch To CardView
    Delete Repo on CardView  busybox
    # Verify
    Delete Success  busybox
    Close Browser

Test Case - Delete Multi Member
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user016  Test1@34    
    Create An New Project  project${d}
    Go Into Project  project${d}
    Switch To Member
    Add Guest Member To Project  user017
    Add Guest Member To Project  user018
    Multi-delete Member  user017  user018
    Delete Success  user017  user018
    Close Browser

Test Case - Project Admin Operate Labels
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user019  Test1@34
    Create An New Project  project${d}
    Go Into Project  project${d}
    Sleep  2
    # Add labels
    Switch To Project Label
    Create New Labels  label_${d}
    Sleep  2
    Update A Label  label_${d}
    Sleep  2
    Delete A Label  label_${d}
    Close Browser

Test Case - Project Admin Add Labels To Repo
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user020  Test1@34
    Create An New Project  project${d}
    Push Image With Tag  ${ip}  user020  Test1@34  project${d}  redis  3.2.10-alpine  3.2.10-alpine
    Push Image With Tag  ${ip}  user020  Test1@34  project${d}  redis  4.0.7-alpine  4.0.7-alpine

    Go Into Project  project${d}
    Sleep  2
    # Add labels
    Switch To Project Label
    Create New Labels  label111
    Create New Labels  label22
    Sleep  2
    Switch To Project Repo
    Go Into Repo  project${d}/redis
    Add Labels To Tag  3.2.10-alpine  label111
    Add Labels To Tag  4.0.7-alpine  label22
    Filter Labels In Tags  label111  label22
    Close Browser

Test Case - Developer Operate Labels
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user021  Test1@34
    Create An New Project  project${d}
    Logout Harbor
    
    Manage Project Member  user021  Test1@34  project${d}  user022  Add
    Change User Role In Project  user021  Test1@34  project${d}  user022  Developer

    Sign In Harbor  ${HARBOR_URL}  user022  Test1@34
    Go Into Project  project${d}
    Sleep  3
    Page Should Not Contain Element  xpath=//a[contains(.,'Labels')]
    Close Browser

Test Case - Scan A Tag In The Repo
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user023  Test1@34
    Create An New Project  project${d}

    Go Into Project  project${d}    
    Push Image  ${ip}  user023  Test1@34  project${d}  hello-world
    Go Into Project  project${d}
    Go Into Repo  project${d}/hello-world
    Scan Repo  latest  Succeed
    Summary Chart Should Display  latest
    Pull Image  ${ip}  user023  Test1@34  project${d}  hello-world
    # Edit Repo Info
    Close Browser

Test Case - Scan As An Unprivileged User
    Init Chrome Driver
    ${d}=    get current date    result_format=%m%s
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  hello-world
 
    Sign In Harbor  ${HARBOR_URL}  user024  Test1@34
    Go Into Project  library
    Go Into Repo  hello-world
    Select Object  latest
    Scan Is Disabled
    Close Browser

Test Case - Scan Image With Empty Vul
    Init Chrome Driver
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  busybox
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  library
    Go Into Repo  busybox
    Scan Repo  latest  Succeed
    Move To Summary Chart
    Wait Until Page Contains  Unknow
    Close Browser

Test Case - Manual Scan All
    Init Chrome Driver
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  redis
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Go To Vulnerability Config
    Trigger Scan Now
    Back To Projects
    Go Into Project  library
    Go Into Repo  redis
    Summary Chart Should Display  latest
    Close Browser

Test Case - Scan Image On Push
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Go Into Project  library
    Goto Project Config
    Enable Scan On Push
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  memcached
    Back To Projects
    Go Into Project  library
    Go Into Repo  memcached
    Summary Chart Should Display  latest
    Close Browser

Test Case - View Scan Results
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user025  Test1@34
    Create An New Project  project${d}    
    Push Image  ${ip}  user025  Test1@34  project${d}  tomcat
    Go Into Project  project${d}
    Go Into Repo  project${d}/tomcat
    Scan Repo  latest  Succeed
    Summary Chart Should Display  latest
    View Repo Scan Details
    Close Browser

Test Case - View Scan Error
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user026  Test1@34
    Create An New Project  project${d}   
    Push Image  ${ip}  user026  Test1@34  project${d}  vmware/photon:1.0
    Go Into Project  project${d}
    Go Into Repo  project${d}/vmware/photon
    Scan Repo  1.0  Fail
    View Scan Error Log
    Close Browser

Test Case - Project Level Image Serverity Policy
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    ${d}=  get current date  result_format=%m%s
    Create An New Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  haproxy
    Go Into Project  project${d}
    Go Into Repo  haproxy
    Scan Repo  latest  Succeed
    Back To Projects
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  0
    Cannot pull image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  haproxy
    Close Browser

Test Case - List Helm Charts
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user027  Test1@34
    Create An New Project  project${d}
    Go Into Project  project${d}
    Sleep  2
    
    Switch To Project Charts
    Upload Chart files
    Go Into Chart Version  ${prometheus_chart_name}
    Wait Until Page Contains  ${prometheus_chart_version}
    Go Into Chart Detail  ${prometheus_chart_version}

    # Summary tab
    Page Should Contain Element  ${summary_markdown}
    Page Should Contain Element  ${summary_container}

    # Dependency tab
    Click Element  xpath=${detail_dependency}
    Sleep  1
    Page Should Contain Element  ${dependency_content}

    # Values tab
    Click Element  xpath=${detail_value}
    Sleep  1
    Page Should Contain Element  ${value_content}

    Go Back To Versions And Delete
    Close Browser

Test Case - Admin Push Signed Image
    Enable Notary Client

    ${rc}  ${output}=  Run And Return Rc And Output  docker pull hello-world:latest
    Log  ${output}

    Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  hello-world:latest
    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group0-Util/notary-push-image.sh ${ip} ${notaryServerEndpoint}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0

    ${rc}  ${output}=  Run And Return Rc And Output  curl -u admin:Harbor12345 -s --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/repositories/library/tomcat/signatures"
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  sha256