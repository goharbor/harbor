*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Suite Setup  Install Harbor to Test Server
Default Tags  BAT

*** Test Cases ***
Test Case - Create An New User
    Start Selenium Standalone Server Locally
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  username=test${d}  email=test${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest

Test Case - Sign With Admin
    Sign In Harbor  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}

Test Case - Create An New Project
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Create An New Project  test${d}
