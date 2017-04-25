*** Settings ***
Documentation  It's an demo case to test robot and drone.
Resource  ../../resources/Harbor-Util.robot
Suite Setup  Install Harbor To Test Server

*** Test Cases ***
Install Harbor to Test Server and add user.
    ${rc}  ${output}=  Run And Return Rc And Output  Log Into Harbor
    Should Be Equal As Integers  ${rc}  0
