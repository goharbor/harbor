*** Settings ***
Documentation  It's an demo case to test robot and drone.
Resource  ../resources/Docker-Util.robot
Default Tags  regression

*** Test Cases ***
Install Harbor to Test Server and add user
    Run Keywords  Start Docker Daemon Locally
    ${rc}  ${output}=  Run And Return Rc And Output  make install GOBUILDIMAGE=golang:1.7.3 COMPILETAG=compile_golangimage CLARITYIMAGE=vmware/harbor-clarity-ui-builder:0.8.4 NOTARYFLAG=true HTTPPROXY=
    Log To Console  \n${rc}
    Log To Console  \n${output}
