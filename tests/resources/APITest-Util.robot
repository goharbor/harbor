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
    [Arguments]  ${testcase_name}  &{param}
    ${current_dir}=  Run  pwd
    ${prev_lvl}  Set Log Level  NONE
    ${param_str}=  Set Variable
    IF  &{param} != {}
        FOR  ${key}  IN  @{param.keys()}
            ${param_str}=  Set Variable  ${param_str} ${key}=${param['${key}']}
        END
    END
    ${rc}  ${output}=  Run And Return Rc And Output  SWAGGER_CLIENT_PATH=${current_dir}/harborclient HARBOR_HOST=${ip} DOCKER_USER=${DOCKER_USER} DOCKER_PWD=${DOCKER_PWD} ${param_str} python ${testcase_name}
    ${prev_lvl}  Set Log Level  ${prev_lvl}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0