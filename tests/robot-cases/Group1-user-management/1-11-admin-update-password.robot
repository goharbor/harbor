*** Settings ***
Documentation  It's an demo case to deploy Harbor with Drone.
Resource  ../../resources/Util.robot
Suite Setup  Start Docker Daemon Locally
Default Tags  regression

*** Test Cases ***
Test Case - Admin Update Password
    Init Chrome Driver
    Sign In Harbor  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Change Password  %{HARBOR_PASSWORD}  Test12#4
    Logout Harbor
    Sign In Harbor  %{HARBOR_ADMIN}  Test12#4
    Close Browser
