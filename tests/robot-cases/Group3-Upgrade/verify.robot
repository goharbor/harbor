*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${ip}
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Upgrade Verify
    ${data}=  Load Json From File  ${CURDIR}${/}data.json
    Run Keyword  Verify User  ${data}
    Run Keyword  Verify Project  ${data}
    Run Keyword  Verify Member Exist  ${data}
    Run Keyword  Verify User System Admin Role  ${data}
    Run Keyword  Verify Endpoint  ${data}
    Run Keyword  Verify Replicationrule  ${data}
    Run Keyword  Verify Project Setting  ${data}
    Run Keyword  Verify System Setting  ${data}
    Run Keyword  Verify Image Tag  ${data}