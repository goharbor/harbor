*** Settings ***
Resource  ../../resources/Util.robot
Suite setup Start Docker Daemon Locally
default tags regression

*** Test Cases ***
Test Case - Edit Email Settings
    Init Chrome Driver
    Sign In Harbor  admin  Harbor12345
    Click Element  xpath=//config//ul/li[3]
    Input Text  xpath=//input[@id="mailServer"]  smtp.vmware.com
    Input Text  xpath=//input[@id="emailPort"]  25
    Input Text  xpath=//input[@id="emailUsername"]  example@vmware.com 
    Input Text  xpath=//input[@id="emailPassword"]  example
    Input Text  xpath=//input[@id="emailFrom"]  example<example@vmware.com>
    #checkbox status by default it is checked
    #Unselect Checkbox  xpath=//input[@id="clr-checkbox-emailSSL"]
    Mouse Down  xpath=//input[@id="clr-checkbox-emailSSL"]
    Mouse Up  xpath=//input[@id="clr-checkbox-emailSSL"]
    Click Button  xpath=//config//div/button[1]

    Logout Harbor
    Sign In Harbor  admin  Harbor12345
    Click Element  xpath=//config//ul/li[3]
    #check value
    Textfield Value Should Be  xpath=//input[@id="mailServer"]  smtp.vmware.com
    Textfield Value Should Be  xpath=//input[@id="emailPort"]  25
    Textfield Value Should Be  xpath=//input[@id="emailUsername"]  example@vmware.com
    #Textfield Value Should Be  xpath=//input[@id="emailPassword"]  example
    Textfield Value Should Be  xpath=//input[@id="emailFrom"]  example<example@vmware.com>
    Checkbox Should Not Be Selected  xpath=//input[@id="clr-checkbox-emailSSL"]
    
    #restore setting
    Input Text  xpath=//input[@id="mailServer"]  smtp.example.com
    Input Text  xpath=//input[@id="emailPort"]  25
    Input Text  xpath=//input[@id="emailUsername"]  example@example.com 
    Input Text  xpath=//input[@id="emailPassword"]  example
    Input Text  xpath=//input[@id="emailFrom"]  example<example@example.com>
    Mouse Down  xpath=//input[@id="clr-checkbox-emailSSL"]
    Mouse Up  xpath=//input[@id="clr-checkbox-emailSSL"]
    Click Button  xpath=//config//div/button[1]
    Close Browser
