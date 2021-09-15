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
    Retry Wait Until Page Contains  ${prometheus_chart_name}
    ${harbor_file_path}  Set Variable  ${current_dir}/${harbor_chart_filename}
    ${harbor_prov_file_path}  Set Variable  ${current_dir}/${harbor_chart_prov_filename}
    Choose File  xpath=${chart_file_browse}  ${harbor_file_path}
    Choose File  xpath=${chart_prov_browse}  ${harbor_prov_file_path}
    Retry Double Keywords When Error  Retry Element Click  xpath=${upload_action_button}  Retry Wait Until Page Not Contains Element  xpath=${upload_action_button}
    Retry Wait Until Page Contains  ${harbor_chart_name}

Go Into Chart Version
    [Arguments]  ${chart_name}
    Retry Element Click  xpath=//hbr-helm-chart//a[contains(., '${chart_name}')]
    Sleep  3

Go Into Chart Detail
    [Arguments]  ${version_name}
    Retry Element Click  xpath=//hbr-helm-chart-version//a[contains(., '${version_name}')]
    Retry Wait Until Page Contains Element  ${chart_detail}

Download Chart File
    [Arguments]  ${chart_name}  ${chart_filename}
    Switch To Project Charts
    ${out}  Run Keyword And Ignore Error  OperatingSystem.File Should Not Exist  ${download_directory}/${chart_filename}
    Run Keyword If  '${out[0]}'=='FAIL'  Run  rm -rf ${download_directory}/${chart_filename}
    Retry File Should Not Exist  ${download_directory}/${chart_filename}
    Retry Element Click  //clr-dg-row[contains(.,'${chart_name}')]//label
    Retry Double Keywords When Error  Retry Element Click  ${download_chart_button}  Retry File Should Exist  ${download_directory}/${chart_filename}
    Retry Element Click  //clr-dg-row[contains(.,'${chart_name}')]//label
    
Multi-delete Chart Files
    [Arguments]    @{obj}
    Switch To Project Charts
    FOR  ${obj}  IN  @{obj}
        Retry Element Click  //clr-dg-row[contains(.,'${obj}')]//label
    END
    #Retry Element Click  xpath=${version_checkbox}
    Retry Double Keywords When Error  Retry Element Click  xpath=${version_delete}  Retry Wait Until Page Contains Element  ${version_confirm_delete}
    Retry Double Keywords When Error  Retry Element Click  ${version_confirm_delete}  Retry Wait Until Page Not Contains Element  xpath=${version_confirm_delete}
    Retry Wait Element  xpath=//clr-dg-placeholder[contains(.,\"We couldn\'t find any charts!\")]

