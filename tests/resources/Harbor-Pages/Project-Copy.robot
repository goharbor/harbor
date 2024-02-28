Documentation  This resource provides any keywords related to the Harbor private registry appliance
Resource  ../../resources/Util.robot

*** Keywords ***
Copy Image
    [Arguments]  ${tag}  ${projectname}  ${reponame}  ${is_success}=${true}
    Retry Element Click  xpath=//clr-dg-row[contains(.,'${tag}')]//label[contains(@class,'clr-control-label')]
    Retry Action Keyword  Copy Image Action  ${projectname}  ${reponame}  ${is_success}

Copy Image Action
    [Arguments]  ${projectname}  ${reponame}  ${is_success}=${true}
    Retry Element Click  ${artifact_action_xpath}
    Retry Element Click  ${artifact_action_copy_xpath}
    Retry Text Input  ${copy_project_name_xpath}  ${projectname}
    Retry Text Input  ${copy_repo_name_xpath}  ${reponame}
    Retry Double Keywords When Error  Retry Element Click  ${confirm_btn}  Wait Until Element Is Not Visible  ${confirm_btn}
    Run Keyword If  '${is_success}' == '${true}'  Retry Wait Until Page Contains  Copy artifact successfully
