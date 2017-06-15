*** Settings ***
resource ../../resources/Util.robot
suite setup Start Docker Daemon Locally
default tags regression

*** Test cases ***
Test Case - Edit Project Creation
	# create normal user and login
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
	#check project creation
    Page Should Contain Element  xpath=//project//div[@class="option-left"]/button
	#logout and login admin
    Logout Harbor
    Sign In Harbor  admin  Harbor12345
	#set limit to admin only
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Click Element  xpath=//select[@id="proCreation"]
    Click Element  xpath=//select[@id="proCreation"]//option[@value="adminonly"]
    Click Element  xpath=//config//div/button[1]
	#logout and login normal user
    Logout Harbor
	Sign In Harbor  tester${d}  Test1@34
	#check if can create project
    Page Should Not Contain Element  xpath=//project//div[@class="option-left"]/button
    Close browser
