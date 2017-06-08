*** Settings ***
Resource ../../resources/Util.robot
Suite Setup Start Docker Daemon Locally
Default Tags regression

*** Test Cases ***
Test Case - Edit Verify Remote Cert
    Init Chrome Driver
    Sign In Harbor user=admin pw=Harbor12345
    click element xpath=//config//ul/li[2]
    #by defalut verify is on
    unselect checkbox xpath=//input[@id="clr-checkbox-verifyRemoteCert"]
    click element xpath=//div/button[1]
    #assume checkbox uncheck
    Logout Harbor
    Sign In Harbor user=admin pw=Harbor12345
    click element xpath=//config//ul/li[2]
    checkbox should not be selected xpath=//input[@id="clr-checkbox-verifyRemoteCert"] 
    close browser
