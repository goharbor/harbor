*** Variables ***
${JFROG_USER}  ${EMPTY}
${JFROG_PWD}  ${EMPTY}
${JFROG_URL}  ${EMPTY}
${JFROG_NAMESPACE}  ${EMPTY}
${OPENAPI_GENERATOR_CLI_URL_DEFAULT}  https://repo1.maven.org/maven2/org/openapitools/openapi-generator-cli/4.3.1/openapi-generator-cli-4.3.1.jar

*** Keywords ***
Make Swagger Client
    ${rc}  ${output}=  Run And Return Rc And Output  pip uninstall setuptools -y
    LogAll  ${output}
    ${rc}  ${output}=  Run And Return Rc And Output  pip install -U pip setuptools
    LogAll  ${output}
    ${openapi_url}=  Get Environment Variable  OPENAPI_GENERATOR_CLI_URL  ${OPENAPI_GENERATOR_CLI_URL_DEFAULT}
    ${rc}  ${output}=  Run And Return Rc And Output  OPENAPI_GENERATOR_CLI_URL=${openapi_url} make swagger_client
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
    ${rc}  ${output}=  Run And Return Rc And Output  SWAGGER_CLIENT_PATH=${current_dir}/harborclient HARBOR_HOST=${ip} DOCKER_USER=${DOCKER_USER} DOCKER_PWD=${DOCKER_PWD} JFROG_USER=${JFROG_USER} JFROG_PWD=${JFROG_PWD} JFROG_URL=${JFROG_URL} JFROG_NAMESPACE=${JFROG_NAMESPACE} ${param_str} python ${testcase_name}
    ${prev_lvl}  Set Log Level  ${prev_lvl}
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0