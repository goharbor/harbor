*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${ip}
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Upgrade Verify
    [Tags]  1.8-latest
    ${data}=  Load Json From File  ${CURDIR}${/}data.json
    Log To Console  "Verify User..."
    Run Keyword  Verify User  ${data}
    Log To Console  "Verify Project..."
    Run Keyword  Verify Project  ${data}
    Log To Console  "Verify Member Exist..."
    Run Keyword  Verify Member Exist  ${data}
    #Run Keyword  Verify Robot Account Exist  ${data}
    Log To Console  "Verify User System Admin Role..."
    Run Keyword  Verify User System Admin Role  ${data}
    Log To Console  "Verify Endpoint..."
    Run Keyword  Verify Endpoint  ${data}
    Log To Console  "Verify Replicationrule..."
    Run Keyword  Verify Replicationrule  ${data}
    Log To Console  "Verify Project Setting..."
    Run Keyword  Verify Project Setting  ${data}
    Log To Console  "Verify System Setting..."
    Run Keyword  Verify System Setting  ${data}
    Log To Console  "Verify Image Tag..."
    Run Keyword  Verify Image Tag  ${data}

Test Case - Upgrade Verify
    [Tags]  1.9-latest
    ${data}=  Load Json From File  ${CURDIR}${/}data.json
    Run Keyword  Verify User  ${data}
    Run Keyword  Verify Project  ${data}
    Run Keyword  Verify Member Exist  ${data}
    #Run Keyword  Verify Robot Account Exist  ${data}
    #Run Keyword  Verify Project-level Whitelist  ${data}
    #Run Keyword  Verify Webhook  ${data}
    Run Keyword  Verify User System Admin Role  ${data}
    Run Keyword  Verify Endpoint  ${data}
    Run Keyword  Verify Replicationrule  ${data}
    Run Keyword  Verify Project Setting  ${data}
    Run Keyword  Verify System Setting  ${data}
    #Run Keyword  Verify System Setting Whitelist  ${data}
    Run Keyword  Verify Image Tag  ${data}