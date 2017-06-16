*** Settings ***
Documentation  It's an demo case to deploy Harbor with Drone.
Resource  ../../resources/Util.robot
Suite Setup  Start Docker Daemon Locally
Default Tags  regression

*** Test Cases ***
Test Case - Assign Sys Admin
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Logout Harbor
    Sign In Harbor  admin  Harbor12345
    Switch to User Tag
    Assign User Admin  tester${d}
    Logout Harbor
    Sign In Harbor  tester${d}  Test1@34
    Administration Tag Should Display
    Close Browser
