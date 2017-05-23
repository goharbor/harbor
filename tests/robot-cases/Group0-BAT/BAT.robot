*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Suite Setup  Install Harbor to Test Server
Default Tags  BAT

*** Test Cases ***
Test Case - Create An New User
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Close Browser

Test Case - Sign With Admin
    Init Chrome Driver
    Sign In Harbor  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Close Browser

Test Case - Create An New Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Create An New Project  test${d}
    Close Browser
