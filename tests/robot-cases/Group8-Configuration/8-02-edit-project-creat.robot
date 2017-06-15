*** Settings ***
resource ../../resources/Util.robot
suite setup Start Docker Daemon Locally
default tags regression

*** Test cases ***
Test Case - Edit Project Creation
# create normal user and login
    Init Chrome Driver
    ${d} = get current date result_format=%m%s
    Create An New User username=test${d} email=test${d}@vmware.com realname=test${d} password=Test1@34 comment=harbor
    Sign In Harbor user=test${d} pw=Test1@34
#check project creation
    should exist xpath=//project//div[@class="option-left"]/button
#logout and login admin
    Logout Harbor
    Sign In Harbor user=admin pw=Harbor12345
#set limit to admin only
    click element xpath=//clr-main-container//nav//ul/li[3]
    click element xpath=//select[@id="proCreation"]
    click element xpath=//select[@id="proCreation"]//option[@value="adminonly"]
    click element xpath=//config//div/button[1]
#logout and login normal user
    Logout Harbor
    Sign In Harbor user=test${d} pw=Test1@34
#check if can create project
    should not exist xpath=//project//div[@class="option-left"]/button
    close browser
