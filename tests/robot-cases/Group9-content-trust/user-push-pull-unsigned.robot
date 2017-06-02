*** settings ***
Resource ../../resources/Util.robot
Suite Setup Start Docker Daemon Locally
Default Tags Regression

*** test cases ***
    ${d}=get current date result_format=%m%s
    ${rc} ${output} =run and return rc and output ip a s eth0|grep inet |awk '{print $2}'|awk -F "/" '{print $1}'
    log ${output}
    ${ip}=${output}

    Create An New User username=unota${d} email=unota{d}@vmware.com  realname=harbortest  newPasword=Test1@34  comment=harbor

    Sign In Harbor user=unota${d} pw=Test1@34
    Create A Project unsigned${d}

    ${rc}${output}= run and return rc and output unset DOCKER_CONTENT_TRUST

    ${rc}${output}=Run And return Rc and Output docker login -u unota${d} -p Test1@34  ${ip}

    ${rc}${output}=run and return rc and output docker tag hello-world ${ip}/unsigned${d}/hello-world:v1
    ${rc}${output}=run and return rc and output docker push ${ip}/unsigned${d}/hello-world:v1
    should be equal as integers ${rc} 0

    #view project
    click element xpath=//a[contains(.,'unsigned')]
    click element xpath=//a[contains(.,'unsigned')]
    sleep 2
    should exist xpath=//clr-icon[@shape="close"]

    #set notary
    ${rc} ${output}= Run And Return Rc And Output export DOCKER_CONTENT_TRUST=1
    ${rc} ${output}= Run And Return Rc And Output export DOCKER_CONTENT_TRUST_SERVER=https://${ip}:4443
    #pull unsigned
    ${rc} ${output}=run and return rc and output docker pull ${ip}/unsigned${d}/hello-world:v1
    
    should not be equal as integers ${rc} 0
