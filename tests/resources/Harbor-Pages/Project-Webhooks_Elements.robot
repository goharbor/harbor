*** Settings ***
Documentation  This resource provides any keywords related to the Harbor private registry appliance

*** Variables ***
${new_webhook_button_xpath}     xpath=//*[@id='new-webhook']
${webhook_name_xpath}           xpath=//*[@id='name']
${webhook_endpoint_id_xpath}    xpath=//*[@id='edit_endpoint_url']
${webhook_auth_header_xpath}    xpath=//*[@id='auth_header']
${action_webhook_xpath}         xpath=//*[@id='action-webhook']
${action_webhook_edit_button}    xpath=//*[@id='edit-webhook']
${action_webhook_disable_or_enable_button}   xpath=//*[@id='toggle-webhook']
${action_webhook_delete_button}  xpath=//*[@id='delete-webhook']
${dialog_delete_button}    xpath=//clr-modal//button[contains(.,'DELETE')]


${create_webhooks_continue_button_xpath}    xpath=//*[@id='new-webhook-continue']
${edit_webhooks_cancel_button_xpath}    xpath=//*[@id='edit-webhook-cancel']
${edit_webhooks_save_button_xpath}    xpath=//*[@id='edit-webhook-save']
${project_webhook_edit_id_xpath}    xpath=//*[@id='edit-webhook']
${project_webhook_enable_id_xpath}    xpath=//*[@id='enable-webhook-action']
${project_webhook_disable_id_xpath}    xpath=//*[@id='disable-webhook-action']
${dialog_disable_id_xpath}    xpath=//*[@id='dialog-action-disable']
${dialog_enable_id_xpath}    xpath=//*[@id='dialog-action-enable']
