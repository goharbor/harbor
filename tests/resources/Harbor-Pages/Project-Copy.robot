Documentation  This resource provides any keywords related to the Harbor private registry appliance
Resource  ../../resources/Util.robot

*** Keywords ***

Copy Image
    [Arguments]  ${tag}  ${projectname}  ${reponame}
    Retry Element Click  xpath=//clr-dg-row[contains(.,'${tag}')]//label
    Sleep  1
    Retry Element Click  ${artifact_action_xpath}
    Sleep  1
    Retry Element Click  ${artifact_action_copy_xpath}
    Sleep  1
    #input necessary info
    Retry Text Input  xpath=${copy_project_name_xpath}  ${projectname}
    Retry Text Input  xpath=${copy_repo_name_xpath}  ${reponame}
    Retry Double Keywords When Error  Retry Element Click  ${confirm_btn}  Retry Wait Until Page Not Contains Element  ${confirm_btn}
