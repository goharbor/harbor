*** Settings ***
Documentation  It's an demo case to deploy Harbor with Drone.
Resource  ../../resources/Util.robot
Suite Setup  Start Docker Daemon Locally
Default Tags  regression

*** Test Cases ***
Test Case - Update User Comment
    Init Chrome Driver
    Sign In Harbor  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Update User Comment  Test12#4
    Logout Harbor
