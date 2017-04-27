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
    OperatingSystem.File Should Exist  ${dockerd-path}
    ${handle}=  Start Process  ${dockerd-path} &>/dev/null &  shell=True
    Log To Console  \n${handle}
    Sleep  5s
    ${output}=  Run  docker pull hello-world
    Log To Console  \n${output}
