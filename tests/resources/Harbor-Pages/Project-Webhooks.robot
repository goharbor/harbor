*** Settings ***
Documentation  Harbor Webhooks
Resource  ../../resources/Util.robot

*** Variables ***

*** Keywords ***
Switch To Project Webhooks
    #Switch To Project Tab Overflow
    Retry Element Click  xpath=//project-detail//a[contains(.,'Webhooks')]
    Sleep  1

Create A New Webhook
    [Arguments]  ${webhook_endpoint_url}  ${auth_header}
    Retry Text Input  ${webhook_endpoint_id_xpath}  ${webhook_endpoint_url}
    Retry Text Input  ${webhook_auth_header_xpath}  ${auth_header}

    Retry Double Keywords When Error  Retry Element Click  ${create_webhooks_continue_button_xpath}  Retry Wait Until Page Not Contains Element  ${create_webhooks_continue_button_xpath}
    Capture Page Screenshot
    Retry Wait Until Page Contains  ${webhook_endpoint_url}

Update A Webhook
    [Arguments]  ${webhook_endpoint_url}  ${auth_header}
    # Cancel input
    Retry Element Click  ${project_webhook_edit_id_xpath}
    Retry Wait Until Page Contains Element  ${webhook_endpoint_id_xpath}
    Input Text  ${webhook_endpoint_id_xpath}  ${webhook_endpoint_url}
    Input Text  ${webhook_auth_header_xpath}  ${auth_header}
    Retry Double Keywords When Error  Retry Element Click  ${edit_webhooks_cancel_button_xpath}  Retry Wait Until Page Not Contains Element  ${edit_webhooks_cancel_button_xpath}
    # Confirm input
    Retry Element Click  ${project_webhook_edit_id_xpath}
    Input Text  ${webhook_endpoint_id_xpath}  ${webhook_endpoint_url}
    Input Text  ${webhook_auth_header_xpath}  ${auth_header}
    Retry Double Keywords When Error  Retry Element Click  ${edit_webhooks_save_button_xpath}  Retry Wait Until Page Not Contains Element  ${edit_webhooks_save_button_xpath}
    Retry Wait Until Page Contains  ${webhook_endpoint_url}
    Capture Page Screenshot

Toggle Enable/Disable State of Same Webhook
    Retry Element Click  ${project_webhook_disable_id_xpath}
    Retry Wait Until Page Contains Element  ${dialog_disable_id_xpath}
    Retry Element Click  ${dialog_disable_id_xpath}
    Retry Wait Until Page Contains Element  ${project_webhook_enable_id_xpath}
    Retry Element Click  ${project_webhook_enable_id_xpath}
    Retry Wait Until Page Contains Element  ${dialog_enable_id_xpath}
    Retry Element Click  ${dialog_enable_id_xpath}
    Retry Wait Until Page Contains Element  ${project_webhook_disable_id_xpath}