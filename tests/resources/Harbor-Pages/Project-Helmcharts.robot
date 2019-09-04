*** Settings ***
Documentation  This resource provides any keywords related to the Harbor private registry appliance
Resource  ../../resources/Util.robot

*** Keywords ***

Switch To Project Charts
    Retry Element Click  ${project_chart_tabpage}
    Retry Wait Until Page Contains Element  ${project_chart_list}

Upload Chart files
    ${current_dir}=  Run  pwd
    Run  wget ${harbor_chart_file_url}
    Run  wget ${harbor_chart_prov_file_url}
    Run  wget ${prometheus_chart_file_url}

    Retry Double Keywords When Error  Retry Element Click  xpath=${upload_chart_button}  Retry Wait Until Page Contains Element  xpath=${upload_action_button}
    ${prometheus_file_path}  Set Variable  ${current_dir}/${prometheus_chart_filename}
    Choose File  xpath=${chart_file_browse}  ${prometheus_file_path}
    Retry Double Keywords When Error  Retry Element Click  xpath=${upload_action_button}  Retry Wait Until Page Not Contains Element  xpath=${upload_action_button}
    #Retry Double Keywords When Error  Retry Element Click  xpath=${upload_action_button}  Retry Wait Until Page Contains Element  xpath=${upload_action_button}
    Retry Wait Until Page Contains  ${prometheus_chart_name}
    Capture Page Screenshot
    ${harbor_file_path}  Set Variable  ${current_dir}/${harbor_chart_filename}
    ${harbor_prov_file_path}  Set Variable  ${current_dir}/${harbor_chart_prov_filename}
    Choose File  xpath=${chart_file_browse}  ${harbor_file_path}
    Choose File  xpath=${chart_prov_browse}  ${harbor_prov_file_path}
    Retry Double Keywords When Error  Retry Element Click  xpath=${upload_action_button}  Retry Wait Until Page Not Contains Element  xpath=${upload_action_button}
    Retry Wait Until Page Contains  ${harbor_chart_name}
    Capture Page Screenshot

Go Into Chart Version
    [Arguments]  ${chart_name}
    Retry Element Click  xpath=//hbr-helm-chart//a[contains(., '${chart_name}')]
    Sleep  3
    Capture Page Screenshot  viewchartversion.png

Go Into Chart Detail
    [Arguments]  ${version_name}
    Retry Element Click  xpath=//hbr-helm-chart-version//a[contains(., '${version_name}')]
    Retry Wait Until Page Contains Element  ${chart_detail}

Multi-delete Chart Files
    [Arguments]    @{obj}
    Switch To Project Charts
    :For  ${obj}  in  @{obj}
    \    Retry Element Click  //clr-dg-row[contains(.,'${obj}')]//label
    #Retry Element Click  xpath=${version_checkbox}
    Capture Page Screenshot
    Retry Double Keywords When Error  Retry Element Click  xpath=${version_delete}  Retry Wait Until Page Contains Element  ${version_confirm_delete}
    Capture Page Screenshot
    Retry Double Keywords When Error  Retry Element Click  ${version_confirm_delete}  Retry Wait Until Page Not Contains Element  xpath=${version_confirm_delete}
    Retry Wait Element  xpath=//clr-dg-placeholder[contains(.,\"We couldn\'t find any charts!\")]
    Capture Page Screenshot

