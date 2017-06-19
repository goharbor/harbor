*** Settings ***
Resource ../../resources/Util.robot
Suite Setup Start Docker Daemon Locally
Default Tags regression

*** Test Cases ***
Test Case - Edit Verify Remote Cert
    Init Chrome Driver
    Sign In Harbor  admin  Harbor12345
    Click Element  xpath=//config//ul/li[2]
    #by defalut verify is on
    #Unselect Checkbox  xpath=//input[@id="clr-checkbox-verifyRemoteCert"]
    Mouse Down  xpath=//input[@id="clr-checkbox-verifyRemoteCert"]
    Mouse Up  xpath=//input[@id="clr-checkbox-verifyRemoteCert"]
    Click Element  xpath=//div/button[1]
    #assume checkbox uncheck
    Logout Harbor
    Sign In Harbor  admin  Harbor12345
    Click Element  xpath=//config//ul/li[2]
    Checkbox Should Not Be Selected  xpath=//input[@id="clr-checkbox-verifyRemoteCert"] 
    Sleep  1
    #restore setting
    Mouse Down  xpath=//input[@id="clr-checkbox-verifyRemoteCert"]
    Mouse Up  xpath=//input[@id="clr-checkbox-verifyRemoteCert"]
    Click Element  xpath=//div/button[1]
    Sleep  1
    Close Browser
