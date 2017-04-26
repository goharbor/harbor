*** Settings ***
Documentation  It's an demo case to test robot and drone.
Resource  ../resources/Util.robot
Default Tags  regression

*** Variables ***
${dockerd-params}

*** Test Cases ***
Install Harbor to Test Server and add user.
    ${output}=  Run  Start Docker Daemon Locally
    Log  ${output}
    ${rc}  ${output}=  Run docker veresion
    Should Be Equal As Integers  ${rc}  0
