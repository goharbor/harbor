*** Settings ***

Resource  ../../resources/Util.robot

*** Variables ***
${HARBOR_VERSION}  V1.1.1

*** Keywords ***

Goto Project Config
    Sleep  3
    Click Element  //project-detail//ul/li[contains(.,'Configuration')]
    Sleep  2

Click Project Public
    Mouse Down  //hbr-project-policy-config//input[@name='public']
    Mouse Up  //hbr-project-policy-config//input[@name='public']

Click Content Trust
    Mouse Down  //hbr-project-policy-config//input[@name='content-trust']
    Mouse Up  //hbr-project-policy-config//input[@name='content-trust']

Click Prevent Running
    Mouse Down  //hbr-project-policy-config//input[@name='prevent-vulnerability-image']
    Mouse Up  //hbr-project-policy-config//input[@name='prevent-vulnerability-image']

Select Prevent Level
#value NEGLIGIBLE LOW MEDIUM HIGH
    [Arguments]  ${level}
    Click Element  //hbr-project-policy-config//select
    Click Element  //hbr-project-policy-config//select/option[contains(.,'${level}')]

Click Auto Scan
    Mouse Down  //hbr-project-policy-config//input[@name='scan-image-on-push']
    Mouse Up  //hbr-project-policy-config//input[@name='scan-image-on-push']

Save Project Config
    Sleep  1
    Click Element  //hbr-project-policy-config//button[contains(.,'SAVE')]

Public Should Be Selected
    Checkbox Should Be Selected  //hbr-project-policy-config//input[@name='public']

Project Should Be Public
    [Arguments]  ${projectName}
    Page Should Contain Element  //clr-dg-row[contains(.,'${projectName}')]//clr-dg-cell[contains(.,'Public')]

Content Trust Should Be Selected
    Checkbox Should Be Selected  //hbr-project-policy-config//input[@name='content-trust']

Prevent Running Should Be Selected
    Checkbox Should Be Selected  //hbr-project-policy-config//input[@name='prevent-vulnerability-image']

Auto Scan Should Be Selected
    Checkbox Should Be Selected  //hbr-project-policy-config//input[@name='scan-image-on-push']

