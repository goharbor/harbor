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
    ${handle}=  Start Process  ${dockerd-path} ${dockerd-params} >${log} 2>&1  shell=True
    Log To Console  \n${handle}
    Process Should Be Running  ${handle}
    :FOR  ${IDX}  IN RANGE  5
    \   ${pid}=  Run  pidof dockerd
    \   Log To Console  \n${pid}
    \   Run Keyword If  '${pid}' != '${EMPTY}'  Set Test Variable  ${dockerd-pid}  ${pid}
    \   Log To Console  \n${dockerd-pid}
    \   Exit For Loop If  '${pid}' != '${EMPTY}'
    \   Sleep  1s
    Should Not Be Equal  '${dockerd-pid}'  '${EMPTY}'
    ${output}=  Run  docker pull hello-world
    Log To Console  \n${output}
