*** Settings ***
Documentation  It's an demo case to deploy Harbor with Drone.
Resource  ../../resources/Util.robot
Suite Setup  Start Docker Daemon Locally
Default Tags  regression

*** Test Cases ***
Test Case - Update User Comment
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Update User Comment  Test12#4
    Logout Harbor
