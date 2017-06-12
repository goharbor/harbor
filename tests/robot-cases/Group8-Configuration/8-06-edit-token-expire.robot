*** settings ***
resource ../../resources/Util.robot
suite setup Start Docker Daemon Locally
default tags regression

*** Test cases ***
Test Case - Edit Token Expire
    Init Chrome Driver
    Sign In Harbor user=admin pw=Harbor12345
    sleep 1
    click element xpath=//clr-main-container//nav//ul/li[3]
    click button xpath=//config//ul/li[4]
    #by default 30,change to other number
    input text xpath=//input[@id="tokenExpiration"] 20
    click button xpath=//config//div/button[1]
    Logout Harbor
    Sign In Harbor user=admin pw=Harbor12345
    sleep 1
    click element xpath=//clr-main-container//nav//ul/li[3]
    click button xpath=//config//ul/li[4]
    Textfield value should be  xpath=//input[@id="tokenExpiration"] 20
    close browser
