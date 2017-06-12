*** Settings ***
Resource ../../resources/Uitl.robot
suite setup Start Docker Daemon Locally
default tags regression

*** Test Cases ***
Test Case - Edit authentication
    Init Chrome Driver
    ${d} = get current date result_format=%m%s
    Sign In Harbor user=admin pw=Harbor12345
    click element xpath=//clr-main-container//nav//ul/li[3]
    click element xpath=//select[@id="authMode"]
    click element xpath=//select[@id="authMode"]//option[@value="ldap_auth"]
    sleep 1
    input text xpath=//input[@id="ldapUrl"]
    input text xpath=//input[@id="ldapSearchDN"]
    input text xpath=//input[@id="ldapSearchPwd"]
    input text xpath=//input[@id="ldapUid"]
    #scope keep subtree
    #click save
    click button xpath=//config//div/button[1]
    Logout Harbor
    #check can change back to db
    Sign In Harbor user=admin pw=Harbor12345
    click element xpath=//clr-main-container//nav//ul/li[3]
    should not exist xpath=//select[@disabled='']
    Logout Harbor
    #signin ldap user
    Sign In Harbor user=user001 pw=user001
    Logout Harbor
    #sign in as admin
    Sign In Harbor user=admin pw=Harbor12345
    click element
    should exist xpath=//select[@disabled='']

    #clean database and restart harbor
    Down Harbor
    ${rc} ${output}= run and return rc and output rm -rf /data
    Prepare
    Up Harbor

    Create An New User username=test${d} email=test${d}@vmware.com  realname=test{d} newPassword=Test1@34 comment=harbor
    Sign In Harbor user=admin pw=Harbor12345
    click element xpath=//clr-main-containter//nav//ul/li[3]
    should exist xpath=//select[@disabled='']
    close browser
