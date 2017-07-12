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
	
    Checkbox Should Not Be Selected  xpath=//input[@id="clr-checkbox-selfReg"]
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

Test case add rule use new endpoint

    Init Chrome Driver
    ${d}=  Get current date  result_format=%m%s
    
    #Create An New User  ${HARBOR_URL}  tester${d}  tester${d}@vmware.com  tester${d}  Test1@34  harbor
    #Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  admin  Harbor12345
    Create An New Project  project${d}
    Mouse down  xpath=//clr-dg-row-master[contains(.,'project${d}')]//a
    Mouse up  xpath=//clr-dg-row-master[contains(.,'project${d}')]//a
    Sleep  1
    #click replication
    Click element  xpath=//project-detail//li[4]
    #click add buton
    Click element  xpath=//hbr-replication//button/clr-icon
    sleep  2
    input text  xpath=//input[@id='policy_name']  aa${d}
    input text  xpath=//textarea[@id='policy_description']   aa${d}
    #click enable
    mouse down  xpath=//input[@id='policy_enable']
    mouse up  xpath=//input[@id='policy_enable']
    #click new endpoint checkbox
    mouse down  xpath=//input[@id='check_new']
    mouse up  xpath=//input[@id='check_new']
    #endpoint info
    input text  xpath=//input[@id='destination_name']  bb${d}
    input text  xpath=//input[@id='destination_url']  bb${d}
    input text  xpath=//input[@id='destination_username']  bb${d}
    input text  xpath=//input[@id='destination_password']  bb${d}
    click element  xpath=//create-edit-rule//button[3]
    sleep  2
    # edit
    #select rule
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'aa${d}')]//clr-dg-action-overflow
    #click edit
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'aa${d}')]//clr-dg-action-overflow//button[contains(.,"Edit")]
    #rules can not be edit if enable
    sleep  1
    page should contain element  xpath=//div[@class='alert alert-warning']
    click element  xpath=//create-edit-rule//button[2]
    #disable to edit
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'aa${d}')]//clr-dg-action-overflow
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'aa${d}')]//clr-dg-action-overflow//button[contains(.,"Disable")]
    click element  xpath=//confirmation-dialog//button[2]
    sleep  2
    #select rule
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'aa${d}')]//clr-dg-action-overflow
    #click edit
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'aa${d}')]//clr-dg-action-overflow//button[contains(.,"Edit")]
    sleep  2
    #rule name
    input text  xpath=//input[@id='policy_name']  cc${d}
    #rule description
    input text   xpath=//textarea[@id='policy_description']  cc${d}
    #endpoint info
    mouse down  xpath=//input[@id='check_new']
    mouse up  xpath=//input[@id='check_new']
    sleep  1
    input text  xpath=//input[@id='destination_name']  bc${d}
    input text  xpath=//input[@id='destination_url']  bc${d}
    input text  xpath=//input[@id='destination_username']  bc${d}
    input text  xpath=//input[@id='destination_password']  bc${d}

    #rule enable status this must after edit endpoint or endpoint can not be edit
    mouse down  xpath=//input[@id='policy_enable']
    mouse up  xpath=//input[@id='policy_enable']
    click element  xpath=//create-edit-rule//button[3]
    sleep  2
    #Rule Disable
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'bc${d}')]//clr-dg-action-overflow
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'bc${d}')]//clr-dg-action-overflow//button[contains(.,"Disable")]
    click element  xpath=//confirmation-dialog//button[2]
    sleep  2
    #Rule enable
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'bc${d}')]//clr-dg-action-overflow
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'bc${d}')]//clr-dg-action-overflow//button[contains(.,"Enable")]
    click element  xpath=//confirmation-dialog//button[2]
    sleep  2
    
    #Rule Delete
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'bc${d}')]//clr-dg-action-overflow
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'bc${d}')]//clr-dg-action-overflow//button[contains(.,"Delete")]
    click element  xpath=//confirmation-dialog//button[2]
    #rule can not deleted if enable
    page should contain element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'bc${d}')]

    #Disable to delete
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'bc${d}')]//clr-dg-action-overflow
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'bc${d}')]//clr-dg-action-overflow//button[contains(.,"Disable")]
    click element  xpath=//confirmation-dialog//button[2]
    sleep  2
    #Rule Delete
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'bc${d}')]//clr-dg-action-overflow
    click element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'bc${d}')]//clr-dg-action-overflow//button[contains(.,"Delete")]
    sleep  1
    click element  xpath=//confirmation-dialog//button[2]
    sleep  2
    #rule can  deleted if disable
    page should not contain element  xpath=//hbr-list-replication-rule//clr-dg-row[contains(.,'bc${d}')]
    close browser 

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
