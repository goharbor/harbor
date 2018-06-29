*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Suite Setup  Nightly Test Setup  ${ip}  ${SSH_PWD}  ${HARBOR_PASSWORD}  ${ip1}
Suite Teardown  Collect Nightly Logs  ${ip}  ${SSH_PWD}  ${ip1}
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${ip}
${SSH_USER}  root
${HARBOR_ADMIN}  admin

*** test case ***

Verify Data
    #get version from extenal argument
    Convert To Float  ${version}
    Run Keyword If  ${version}==1.1
       Run Keywords  Verify User  Verify Project  Verify Member Exist  Verify Image Tag  Verify Endpoint  Verify Replicationrule 
       Else If  ${version}==1.2
       Run Keywords  Verify User  Verify Project  Verify Member Exist  Verify Image Tag  Verify Endpoint  Verify Replicationrule
       Else If  ${version}==1.3
       Run Keywords  Verify User  Verify Project  Verify Member Exist  Verify Image Tag  Verify Endpoint  Verify Replicationrule  Verify System Setting  Verify Project Setting
       Else If  ${version}==1.4
       Run Keywords  Verify User  Verify Project  Verify Member Exist  Verify Image Tag  Verify Endpoint  Verify Replicationrule  Verify System Setting  Verify Project Setting
       Else If  ${version}==1.5
       Run Keywords  Verify User  Verify Project  Verify Member Exist  Verify Image Tag  Verify Endpoint  Verify Replicationrule  Verify System Setting  Verify Project Setting  Verify Project Label  Verify Syslabel
       Else
       Log To Consle  "Version Not Supported"

