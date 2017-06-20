*** Settings ***
Resource ../../resources/Util.robot
Suite Setup Start Docker Daemon Locally
Default Tags regression

*** Test Cases ***
Test Case - Edit Self-Registration
#login as admin
    Init Chrome Driver
    Sign In Harbor  admin  Harbor12345
#disable self reg
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    #Unselect Checkbox  xpath=//input[@id="clr-checkbox-selfReg"]
    Mouse Down  xpath=//input[@id="clr-checkbox-selfReg"]
    Mouse Up  xpath=//input[@id="clr-checkbox-selfReg"]
    Click Element  xpath=//div/button[1]
#logout and check
    Logout Harbor
    Page Should Not Conatain Element  xpath=//a[@class="signup"]
    Sign In Harbor  admin  Harbor12345
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Checkbox Should Not Be Selected  xpath=//input[@id="clr-checkbox-selfReg"]
    Sleep  1
    #restore setting
    Mouse Down  xpath=//input[@id="clr-checkbox-selfReg"]
    Mouse Up  xpath=//input[@id="clr-checkbox-selfReg"]
    Click Element  xpath=//div/button[1]
    Close Browser
