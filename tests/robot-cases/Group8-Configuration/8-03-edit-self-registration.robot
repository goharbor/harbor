*** Settings ***
resource ../../resources/Util.robot
suite setup Start Docker Daemon Locally
default tags regression

*** Test Cases ***
Test Case - Edit Self-Registration
#login as admin
    Init Chrome Driver
    Sign In Harbor user=admin pw=Harbor12345
#disable self reg
    click element xpath=//clr-main-container//nav//ul/li[3]
    unselect checkbox xpath=input[@id="clr-checkbox-selfReg"]
#logout and check
    Logout Harbor
    should not exist xpath=//a[@class="signup"]
    Sign In Harbor user=admin pw=Harbor12345
    click element xpath=//clr-main-container//nav//ul/li[3]
    checkbox should not be selected xpath=//input[@id="clr-checkbox-selfReg"]
    Close browser
