*** Settings ***
Documentation  It's an demo case to deploy Harbor with Drone.
Resource  ../../resources/Util.robot
Default Tags  regression

*** Test Cases ***
Install Harbor to Test Server
    Log To Console  \nstart docker
    Run Keywords  Start Docker Daemon Locally
    Log To Console  \ndocker started success, config harbor cfg
    Run Keywords  Config Harbor cfg
    Log To Console  \ncomplile and up harbor now
    Run Keywords  Compile and Up Harbor With Source Code
    ${rc}  ${output}=  Run And Return Rc And Output  docker ps
    Should Be Equal As Integers  ${rc}  0
    Log To Console  \n${output}
    Sleep 10s
    ${rc}  ${output}=  Run And Return Rc And Output  curl -s -L -H "Accept: application/json" http://localhost/
    Log To Console  \n${output}
    Should Be Equal As Integers  ${rc}  0
