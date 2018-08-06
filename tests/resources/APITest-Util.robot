*** Keywords ***
Harbor API Test 
    [Arguments]  ${testcase_name}
    ${rc}  ${output}=  Run And Return Rc And Output  SWAGGER_CLIENT_PATH=./harborclient HARBOR_HOST=${ip} python ${testcase_name} 
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0