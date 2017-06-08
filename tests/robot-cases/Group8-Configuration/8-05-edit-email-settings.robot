*** Settings ***
Resource  ../../resources/Util.robot
Suite setup Start Docker Daemon Locally
default tags regression

*** Test Cases ***
Test Case - Edit Email Settings
    Init Chrome Driver
    Sign In Harbor user=admin pw=Harbor12345
    click element xpath=//config//ul/li[3]
    input text xpath=//input[@id="mailServer"] smtp.vmware.com
    input text xpath=//input[@id="emailPort"] 25
    input text xpath=//input[@id="emailUsername"] example@vmware.com 
    input text xpath=//input[@id="emailPassword"] example
    input text xpath=//input[@id="emailFrom"] example<example@vmware.com>
    #checkbox status by default it is checked
    unselect checkbox xpath=//input[@id="clr-checkbox-emailSSL"]
    click button xpath=//config//div/button[1]

    Logout Harbor
    Sign In Harbor user=admin pw=Harbor12345
    click element xpath=//config//ul/li[3]
    #check value
    textfield value should be xpath=//input[@id="mailServer"] smtp.vmware.com
    textfield value should be xpath=//input[@id="emailPort"] 25
    textfield value should be xpath=//input[@id="emailUsername"] example@vmware.com
    textfield value should be xpath=//input[@id="emailPassword"] example
    textfield value should be xpath=//input[@id="emailFrom"] example<example@vmware.com>
    checkbox should not be selected xpath=//input[@id="clr-checkbox-emailSSL"]
    close browser
