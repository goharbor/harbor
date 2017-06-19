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
    Sign In Harbor  ${HARBOR_URL}  admin  Harbor12345
	#set limit to admin only
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Click Element  xpath=//select[@id="proCreation"]
    Click Element  xpath=//select[@id="proCreation"]//option[@value="adminonly"]
    Click Element  xpath=//config//div/button[1]
	Capture Page Screenshot
	#logout and login normal user
    Logout Harbor
	Sign In Harbor  ${HARBOR_URL}  tester${d}  Test1@34
	#check if can create project
	Capture Page Screenshot
    Page Should Not Contain Element  xpath=//project//div[@class="option-left"]/button
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  admin  Harbor12345
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Click Element  xpath=//select[@id="proCreation"]
    Click Element  xpath=//select[@id="proCreation"]//option[@value="everyone"]
    Click Element  xpath=//config//div/button[1]
    Sleep  2
    Close browser

Test Case - Assign Sys Admin
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=${HARBOR_URL}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  admin  Harbor12345
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
	Capture Page Screenshot
    Wait Until Page Contains  test${d}1
    Wait Until Page Contains  test${d}2
    Wait Until Page Contains  test${d}3
    Close Browser

Test Case - Push Image
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Create An New Project  test${d}
    Logout Harbor

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

    Sign In Harbor  tester${d}  Test1@34
    Sleep  2
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/nav/section/a[2]
    Sleep  2
    Click Element  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/nav/section/a[1]
    Sleep  2
    Click Element  xpath=//harbor-app/harbor-shell/clr-main-container//project//list-project/clr-datagrid//clr-dg-row/clr-dg-row-master/clr-dg-cell[2]/a[contains(.,"test")]
    Sleep  2
    Wait Until Page Contains  test${d}/hello-world
    Sleep  1
    Close Browser

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
