*** Keywords ***
Make Swagger Client
    ${rc}  ${output}=  Run And Return Rc And Output  pip uninstall setuptools -y
    LogAll  ${output}
    ${rc}  ${output}=  Run And Return Rc And Output  pip install -U pip setuptools
    LogAll  ${output}
    ${rc}  ${output}=  Run And Return Rc And Output  make swagger_client
    LogAll  ${output}
    [Return]  ${rc}

Setup API Test
    Retry Keyword N Times When Error  10  Make Swagger Client

Harbor API Test
    [Arguments]  ${testcase_name}
    ${current_dir}=  Run  pwd
    Log To Console  ${current_dir}
    Log To Console  ${ip}
    ${rc}  ${output}=  Run And Return Rc And Output  SWAGGER_CLIENT_PATH=${current_dir}/harborclient HARBOR_HOST=${ip} DOCKER_USER=${DOCKER_USER} DOCKER_PWD=${DOCKER_PWD} python ${testcase_name}
    Log To Console  ${output}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0
