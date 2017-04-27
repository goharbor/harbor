*** Settings ***
Documentation  It's an demo case to test robot and drone.
Library  OperatingSystem
Library  Process
Default Tags  regression

*** Variables ***
${dockerd-params}  dockerd
${dockerd-path}  /usr/local/bin/dockerd-entrypoint.sh
${log}  ./daemon-local.log

*** Test Cases ***
Install Harbor to Test Server and add user
    ${output}=  Run  docker pull hello-world
    Log To Console  \n${output}
    ${rc}  ${output}=  Run And Return Rc And Output  make compile_clarity GOBUILDIMAGE=golang:1.7.3 COMPILETAG=compile_golangimage CLARITYIMAGE=vmware/harbor-clarity-ui-builder:0.8.4 NOTARYFLAG=true HTTPPROXY=
    Log To Console  \n${rc}
    Log To Console  \n${output}
