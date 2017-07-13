*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Suite Setup  Install Harbor to Test Server
Default Tags  BAT

*** Variables ***
${HARBOR_URL}  http://localhost

*** Test Cases ***
Test Case - Create An New User
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Close Browser

Test Case - Sign With Admin
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Close Browser

Test Case - Update User Comment
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Update User Comment  Test12#4
    Logout Harbor

Test Case - Update Password
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Change Password  Test1@34  Test12#4
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  tester${d}  Test12#4
    Close Browser
	
Test Case - Edit Project Creation
	# create normal user and login
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
	
	#check project creation
    Page Should Contain Element  xpath=//project//div[@class="option-left"]/button
	
	#logout and login admin
    Logout Harbor
	Sleep  3
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
	Set Pro Create Admin Only
		
	#logout and login normal user
    Logout Harbor
	Sign In Harbor  ${HARBOR_URL}  tester${d}  Test1@34
	Page Should Not Contain Element  xpath=//project//div[@class="option-left"]/button
    
	Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
	
    Set Pro Create Every One
    Close browser

Test Case - Edit Self-Registration
	#login as admin
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Disable Self Reg
	    
	Logout Harbor
    Page Should Not Contain Element  xpath=//a[@class="signup"]
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
   
	Switch To Configure	
    Self Reg Should Be Disabled
    Sleep  1
    
	#restore setting
    Enable Self Reg
    Close Browser

Test Case - Edit Verify Remote Cert
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}

    Switch To System Replication
    Check Verify Remote Cert

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}

    Switch To System Replication
    Should Verify Remote Cert Be Enabled

    #restore setting
    Check Verify Remote Cert
    Close Browser  

Test Case - Edit Email Settings
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    
	Switch To Email
    Config Email

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    
	Switch To Email
    Verify Email
    
    Close Browser

Test Case - Edit Token Expire
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
	Switch To System Settings
	Modify Token Expiration  20
    Logout Harbor
    
	Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
	Switch To System Settings
    Token Must Be Match  20
    
	#reset to default
    Modify Token Expiration  30
    Close Browser   
	
Test Case - Create An Replication Rule New Endpoint
    Init Chrome Driver
    ${d}=  Get current date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Create An New Project  project${d}
    Sleep  3
	Go Into Project  project${d}
    Switch To Replication
    Create An New Rule With New Endpoint  policy_name=test_policy_${d}  policy_description=test_description  destination_name=test_destination_name_${d}  destination_url=test_destination_url_${d}  destination_username=test_destination_username  destination_password=test_destination_password
	Close Browser 

Test Case-Manage Project Member
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
    ${rc}=  ${output}=  run and return rc and output  docker login -u usera${d} -p Test1@34 ${ip}
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

Test Case - Assign Sys Admin
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

Test Case - Create An New Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Create An New Project  test${d}
    Close Browser

Test Case - User View Projects
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Create An New Project  test${d}1
    Create An New Project  test${d}2
    Create An New Project  test${d}3
    Switch To Log
	Capture Page Screenshot  UserViewProjects.png
    Wait Until Page Contains  test${d}1
    Wait Until Page Contains  test${d}2
    Wait Until Page Contains  test${d}3
    Close Browser

Test Case - Push Image
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Create An New Project  test${d}
    Close Browser

    ${rc}  ${ip}=  Run And Return Rc And Output  ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'
    Log To Console  ${ip}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  docker pull hello-world
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker login -u tester${d} -p Test1@34 ${ip}
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  docker tag hello-world ${ip}/test${d}/hello-world:latest
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc  docker push ${ip}/test${d}/hello-world:latest
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0

    Init Chrome Driver
    Go To    ${HARBOR_URL}
    Sleep  2
    ${title}=  Get Title
    Should Be Equal  ${title}  Harbor
    Sign In Harbor  ${HARBOR_URL}  tester${d}  Test1@34
    Sleep  2
	Capture Page Screenshot  PushImage1.png
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/nav/section/a[2]
    Sleep  2
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/nav/section/a[1]
    Sleep  2
    Mouse Down  xpath=//project//list-project/clr-datagrid//clr-dg-row/clr-dg-row-master//a[contains(.,"test")]
    Mouse Up  xpath=//project//list-project/clr-datagrid//clr-dg-row/clr-dg-row-master//a[contains(.,"test")]
    Sleep  2
    Wait Until Page Contains  test${d}/hello-world

Test Case - Ldap Sign in and out
    Switch To LDAP
    Init Chrome Driver
    ${rc}=  Run And Return Rc  docker pull vmware/harbor-ldap-test:1.1.1
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0
    ${rc}=  Run And Return Rc   docker run --name ldap-container -p 389:389 --detach vmware/harbor-ldap-test:1.1.1
    Log  ${rc}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  docker ps
    Should Be Equal As Integers  ${rc}  0
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Switch To Configure
    Init LDAP
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user001  user001
    Close Browser

Test Case - Admin Push Signed Image
    Switch To Notary

    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group9-Content-trust/notary-push-image.sh
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0

    ${rc}  ${ip}=  Run And Return Rc And Output  ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'
    Log  ${ip}

    ${rc}  ${output}=  Run And Return Rc And Output  curl -u admin:Harbor12345 -s --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/repositories/library/tomcat/signatures"
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  sha256
	
Test Case - Notary Inteceptor
    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group9-Content-trust/notary-pull-image-inteceptor.sh
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0

    Down Harbor  with_notary=true
    ${rc}  ${output}=  Run And Return Rc And Output  echo "PROJECT_CONTENT_TRUST=1\n" >> ./make/common/config/ui/env
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  cat ./make/common/config/ui/env

    Log To Console  ${output}		
	Up Harbor  with_notary=true	
    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group9-Content-trust/notary-pull-image-inteceptor.sh
    Log To Console  ${output}
	
	Down Harbor  with_notary=true
	${rc}  ${output}=  Run And Return Rc And Output  sed "s/^PROJECT_CONTENT_TRUST=1.*/PROJECT_CONTENT_TRUST=0/g" -i ./make/common/config/ui/env
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc And Output  cat ./make/common/config/ui/env
		
	Up Harbor  with_notary=true	
    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group9-Content-trust/notary-pull-image-inteceptor.sh
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0	
