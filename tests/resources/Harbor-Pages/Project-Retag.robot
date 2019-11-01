Documentation  This resource provides any keywords related to the Harbor private registry appliance
Resource  ../../resources/Util.robot

*** Keywords ***

Retag Image
    [Arguments]  ${tag}  ${projectname}  ${reponame}  ${tagname}
    Retry Element Click  xpath=//clr-dg-row[contains(.,'${tag}')]//label
    Retry Element Click  xpath=${retag_btn}
    #input necessary info
    Retry Text Input  xpath=${project-name_xpath}  ${projectname}
    Retry Text Input  xpath=${repo-name_xpath}  ${reponame}
    Retry Text Input  xpath=${tag-name_xpath}  ${tagname}
    Retry Double Keywords When Error  Retry Element Click  ${confirm_btn}  Retry Wait Until Page Not Contains Element  ${confirm_btn}
