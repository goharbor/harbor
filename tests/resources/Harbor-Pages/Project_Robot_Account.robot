*** Settings ***

Resource  ../../resources/Util.robot

*** Variables ***

*** Keywords ***
Switch To Project Robot Account
    #Switch To Project Tab Overflow
    Retry Element Click  ${project_robot_account_tabpage}
    Retry Wait Until Page Contains Element  ${project_robot_account_create_btn}

Create A Project Robot Account
    [Arguments]  ${robot_account_name}  ${expiration_type}  ${description}=${null}  ${days}=${null}  ${resources}=${null}
    Retry Element Click  ${project_robot_account_create_btn}
    Retry Wait Element Should Be Disabled  //button[text()='Next']
    Retry Text Input  ${project_robot_account_create_name_input}    ${robot_account_name}
    Run Keyword If  '${description}' != '${null}'  Retry Text Input  //textarea  ${description}
    Select From List By Value  ${project_robot_account_create_sexpiration_type_btn}  ${expiration_type}
    Run Keyword If  '${expiration_type}' == 'days'  Retry Text Input  ${project_robot_account_token_expiration_days}  ${days}
    Retry Double Keywords When Error  Retry Button Click  //button[text()='Next']  Retry Wait Element Not Visible  //button[text()='Next']
    Retry Wait Element Should Be Disabled  ${project_robot_account_create_finish_btn}
    ${first_resource}=  Set Variable  ${resources}[0]
    ${permission_count}=  Create Dictionary
    ${total}=  Set Variable  0
    IF  '${first_resource}' == 'all'
        Set To Dictionary  ${permission_count}  all=56
        ${total}=  Set Variable  56
        Retry Element Click  //span[text()='Select all']
    ELSE
        FOR  ${item}  IN  @{resources}
            ${elements}=  Get WebElements  //table//tr[./td[text()='${item}']]//label
            ${elements_count}=  Get Length  ${elements}
            Set To Dictionary  ${permission_count}  ${item}=${elements_count}
            ${total}=  Evaluate  ${total} + ${elements_count}
            FOR  ${element}  IN  @{elements}
                Retry Element Click  ${element}
            END
        END
    END
    Retry Double Keywords When Error  Retry Element Click  ${project_robot_account_create_finish_btn}  Retry Wait Until Page Not Contains Element  ${project_robot_account_create_finish_btn}
    ${robot_account_name}=  Get Text  ${project_robot_account_name_xpath}
    ${token}=  Get Value  ${project_robot_account_token_input}
    Retry Element Click  //hbr-copy-input//clr-icon
    IF  '${days}' == '${null}'
        ${expires}=  Set Variable  Never Expires
    ELSE
        ${days}=  Evaluate  ${days} - 1
        ${expires}=  Set Variable  ${days}d 23h
    END
    Retry Wait Element Visible  //clr-dg-row[.//clr-dg-cell[contains(.,'${robot_account_name}')] and .//clr-icon[contains(@class, 'color-green')] and .//button[text()=' ${total} PERMISSION(S) '] and .//span[contains(.,'${expires}')] and .//clr-dg-cell[text()='${description}'] ]
    [Return]  ${robot_account_name}  ${token}  ${permission_count}

Check Project Robot Account Permission
    [Arguments]  ${robot_account_name}  ${permission_count}
    Retry Button Click  //clr-dg-row[.//clr-dg-cell[contains(., '${robot_account_name}')]]//button
    FOR  ${key}  IN  @{permission_count.keys()}
        Retry Wait Element Count  //table//tr[./td[text()=' ${key} ']]//label  ${permission_count['${key}']}
    END
    Retry Double Keywords When Error  Retry Button Click  //button[@aria-label='Close']  Retry Wait Until Page Not Contains Element  //button[@aria-label='Close']

Check Project Robot Account API Permission
    [Arguments]  ${robot_account_name}  ${token}  ${admin_user_name}  ${admin_password}  ${project_id}  ${project_name}  ${source_artifact_name}  ${source_artifact_tag}  ${resources}  ${expected_status}=0
    ${rc}  ${output}=  Run And Return Rc And Output  USER_NAME='${robot_account_name}' PASSWORD='${token}' ADMIN_USER_NAME=${admin_user_name} ADMIN_PASSWORD=${admin_password} HARBOR_BASE_URL=https://${ip}/api/v2.0 PROJECT_ID=${project_id} PROJECT_NAME=${project_name} SOURCE_ARTIFACT_NAME=${source_artifact_name} SOURCE_ARTIFACT_TAG=${source_artifact_tag} RESOURCES=${resources} python ./tests/apitests/python/test_project_permission.py
    Should Be Equal As Integers  ${rc}  ${expected_status}
