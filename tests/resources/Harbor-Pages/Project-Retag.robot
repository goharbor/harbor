Documentation  This resource provides any keywords related to the Harbor private registry appliance
Resource  ../../resources/Util.robot

*** Keywords ***

Retag Image
    [Arguments]  ${tag}  ${projectname}  ${reponame}  ${tagname}
    Click Element  xpath=//clr-dg-row[contains(.,'${tag}')]//label
    Sleep  1
    Click Element  xpath=${retag_btn}
    Sleep  1
    #input necessary info
    Input Text  xpath=${project-name_xpath}  ${projectname}
    Input Text  xpath=${repo-name_xpath}  ${reponame}
    Input Text  xpath=${tag-name_xpath}  ${tagname}
    Click Element  xpath=${confirm_btn}
