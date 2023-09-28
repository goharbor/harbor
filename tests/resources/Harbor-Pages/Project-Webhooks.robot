*** Settings ***
Documentation  Harbor Webhooks
Resource  ../../resources/Util.robot

*** Variables ***

*** Keywords ***
Switch To Project Webhooks
    Retry Element Click  xpath=//project-detail//a[contains(.,'Webhooks')]

Create A New Webhook
    [Arguments]  ${webhook_name}  ${webhook_endpoint_url}  ${payload_format}=Default  ${event_type}=@{EMPTY}
    Retry Element Click  ${new_webhook_button_xpath}
    Retry Text Input  ${webhook_name_xpath}  ${webhook_name}
    Retry Text Input  ${webhook_endpoint_id_xpath}  ${webhook_endpoint_url}
    Run Keyword If  '${payload_format}' != 'Default'  Select Payload Format  ${payload_format}
    ${len}=  Get Length  ${event_type}
    Run Keyword If  ${len} > 0  Select Event Type  @{event_type}
    Retry Double Keywords When Error  Retry Element Click  ${create_webhooks_continue_button_xpath}  Retry Wait Until Page Not Contains Element  ${create_webhooks_continue_button_xpath}
    Retry Wait Until Page Contains  ${webhook_name}

Select Payload Format
    [Arguments]  ${payload_format}
    Retry Double Keywords When Error  Retry Element Click  ${webhook_payload_format_xpath}  Retry Element Click  ${webhook_payload_format_xpath}//option[@value='${payload_format}']

Select Event Type
    [Arguments]  @{event_type}
    ${elements}=  Get WebElements  //form//div[contains(@class,'clr-control-inline')]//label[contains(@class,'clr-control-label')]
    FOR  ${element}  IN  @{elements}
        Retry Element Click  ${element}
    END
    FOR  ${element}  IN  @{event_type}
        Retry Element Click  //form//div[contains(@class,'clr-control-inline')]//label[contains(@class,'clr-control-label') and contains(.,'${element}')]
    END

Update A Webhook
    [Arguments]  ${old_webhook_name}  ${new_webhook_name}  ${new_webhook_enpoint}  ${payload_format}=Default
    # select one webhook
    Retry Element Click   xpath=//clr-dg-row[contains(.,'${old_webhook_name}')]//div[contains(@class,'datagrid-select')]
    Retry Element Click  ${action_webhook_xpath}
    Retry Element Click  ${action_webhook_edit_button}

    #cancel1
    Retry Double Keywords When Error  Retry Element Click  ${edit_webhooks_cancel_button_xpath}  Retry Wait Until Page Not Contains Element  ${edit_webhooks_cancel_button_xpath}
    #confirm
    Retry Element Click  ${action_webhook_xpath}
    Retry Element Click  ${action_webhook_edit_button}
    Retry Text Input   ${webhook_name_xpath}   ${new_webhook_name}
    Retry Text Input  ${webhook_endpoint_id_xpath}  ${new_webhook_enpoint}
    Select Payload Format  ${payload_format}
    Retry Double Keywords When Error  Retry Element Click  ${edit_webhooks_save_button_xpath}  Retry Wait Until Page Not Contains Element  ${edit_webhooks_save_button_xpath}
    Retry Wait Until Page Contains  ${new_webhook_name}

Enable/Deactivate State of Same Webhook
    [Arguments]  ${webhook_name}
    Retry Element Click   xpath=//clr-dg-row[contains(.,'${webhook_name}')]//div[contains(@class,'datagrid-select')]
    Retry Element Click   ${action_webhook_xpath}
    Retry Element Click   ${action_webhook_disable_or_enable_button}
    Retry Wait Until Page Contains Element  ${dialog_disable_id_xpath}
    Retry Element Click  ${dialog_disable_id_xpath}
    # contain deactivated webhook
    Retry Wait Until Page Contains Element   xpath=//clr-dg-row[contains(.,'${webhook_name}')]//span[contains(.,'Deactivated')]

    Retry Element Click   xpath=//clr-dg-row[contains(.,'${webhook_name}')]//div[contains(@class,'datagrid-select')]
    Retry Element Click   ${action_webhook_xpath}
    Retry Element Click   ${action_webhook_disable_or_enable_button}
    Retry Wait Until Page Contains Element  ${dialog_enable_id_xpath}
    Retry Element Click  ${dialog_enable_id_xpath}
    # not contain deactivated webhook
    Retry Wait Until Page Not Contains Element   xpath=//clr-dg-row[contains(.,'${webhook_name}')]//span[contains(.,'Deactivated')]

Delete A Webhook
    [Arguments]  ${webhook_name}
    Retry Element Click   xpath=//clr-dg-row[contains(.,'${webhook_name}')]//div[contains(@class,'datagrid-select')]
    Retry Element Click   ${action_webhook_xpath}
    Retry Element Click   ${action_webhook_delete_button}
    Retry Wait Until Page Contains Element  ${dialog_delete_button}
    Retry Element Click  ${dialog_delete_button}
    Retry Wait Until Page Not Contains Element   xpath=//clr-dg-row[contains(.,'${webhook_name}')]//div[contains(@class,'datagrid-select')]
