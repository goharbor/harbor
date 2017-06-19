*** Settings ***
Resource ../../resources/Util.robot
Suite Setup Start Docker Daemon Locally
Default Tags regression

*** Test Cases ***
Test Case - Edit Token Expire
    Init Chrome Driver
    Sign In Harbor  admin  Harbor12345
    Sleep  1
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Click Button  xpath=//config//ul/li[4]
    #by default 30,change to other number
    Input Text  xpath=//input[@id="tokenExpiration"]  20
    Click Button  xpath=//config//div/button[1]
    Logout Harbor
    Sign In Harbor  admin  Harbor12345
    Sleep  1
    Click Element  xpath=//clr-main-container//nav//ul/li[3]
    Click Button  xpath=//config//ul/li[4]
    Textfield Value Should Be  xpath=//input[@id="tokenExpiration"]  20
    #restore setting
    Input Text  xpath=//input[@id="tokenExpiration"]  30
    Click Button  xpath=//config//div/button[1]
    Close Browser
