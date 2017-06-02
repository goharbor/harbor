*** Settings ***
Resource ../../resources/Util.robot
Suite Setup Start Docker Daemon Locally
Default Tags Regression

*** Test Cases ***
    ${d}= get current date result_format=%m%s
    ${rc} ${output}= run and return rc and output ip a s eth0|grep inet |awk '{print $2}'|awk -F "/" '{print $1}'
    log ${output}
    ${ip}=${output}

    Create An New User username=nota${d} email=nota${d}@vmware.com realname=harbortest newPassword=Test1@34 comment=harbor

    Sign In Harbor user=nota${d} pw=Test1@34

    Create An New Project signed${d}

    ${rc} ${output}= Run And Return Rc And Output export DOCKER_CONTENT_TRUST=1

    ${rc} ${output}= Run And Return Rc And Output export DOCKER_CONTENT_TRUST_SERVER=https://${ip}:4443

    ${rc} ${output}=Run And return Rc and Output docker login -u nota${d} -p Test1@34 ${ip}

    ${rc} ${output}=run and return rc and output docker tag hello-world ${ip}/signed${d}/hello-world:v1

#    run expect script ${ip}
    Run process python ppexpect.py /signed${d}/hello-world:v1

    #view project
    click element xpath=//a[contains(.,"signed")]
    click element xpath=//a[contians(.,"signed")]
    should exist xpath=//clr-icon[@shape="check"]

    #pull signed
    ${rc} ${output}=run and return rc and output docker pull ${ip}/signed${d}/hello-world:v1
    should be equal as integers ${rc} 0

    #delete signed

    click element xpath=//clr-icon[@shape="ellipsis-vertical"]
    click element xpath=//button[contains(.,"Delete")]
    #delete dialog
    should exist xpath=//confiramtion-dialog
    click element xpath=//confiramtion-dialog//button
    #tag shold still exist
    should exist xpath=////clr-icon[@shape="check"]

    #run notary command
    ${rc} ${output}=run and return rc and output notary -s https://${ip}:4443 ~/.docker/trust remove -p ${ip}/signed${d}/hello-world v1

    #should exist cross
    should exist xpath=//clr-icon[@shape="close"]






