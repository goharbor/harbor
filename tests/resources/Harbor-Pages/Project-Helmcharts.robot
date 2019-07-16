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

    Retry Double Keywords When Error  Retry Element Click  xpath=${upload_chart_button}  Retry Wait Until Page Contains Element  xpath=${upload_action_button}
    ${harbor_file_path}  Set Variable  ${current_dir}/${harbor_chart_filename}
    ${harbor_prov_file_path}  Set Variable  ${current_dir}/${harbor_chart_prov_filename}
    Choose File  xpath=${chart_file_browse}  ${harbor_file_path}
    Choose File  xpath=${chart_prov_browse}  ${harbor_prov_file_path}
    Retry Double Keywords When Error  Retry Element Click  xpath=${upload_action_button}  Retry Wait Until Page Not Contains Element  xpath=${upload_action_button}

    Retry Wait Until Page Contains  ${prometheus_chart_name}

Go Into Chart Version
    [Arguments]  ${chart_name}
    Retry Element Click  xpath=//hbr-helm-chart//a[contains(., '${chart_name}')]
    Sleep  3
    Capture Page Screenshot  viewchartversion.png

Go Into Chart Detail
    [Arguments]  ${version_name}
    Retry Element Click  xpath=//hbr-helm-chart-version//a[contains(., '${version_name}')]
    Retry Wait Until Page Contains Element  ${chart_detail}

Go Back To Versions And Delete
    Retry Element Click  xpath=${version_bread_crumbs}
    Retry Element Click  xpath=${version_checkbox}
    Retry Element Click  xpath=${version_delete}
    :For  ${n}  IN RANGE  1  6
    \    Log To Console  Trying Go Back To Versions And Delete ${n} times ...
    \    Retry Element Click  xpath=${version_confirm_delete}
    \    Capture Page Screenshot
    \    ${out}  Run Keyword And Ignore Error  Retry Wait Until Page Contains Element  xpath=${helmchart_content}
    \    Capture Page Screenshot
    \    Log To Console  Return value is ${out[0]}
    \    Exit For Loop If  '${out[0]}'=='PASS'
    \    Sleep  1
