*** Settings ***
Documentation  This resource provides any keywords related to the Harbor private registry appliance

*** Variables ***
${webhook_endpoint_id_xpath}    xpath=//*[@id='edit_endpoint_url']
${webhook_auth_header_xpath}    xpath=//*[@id='auth_header']
${create_webhooks_continue_button_xpath}    xpath=//*[@id='new-webhook-continue']
${edit_webhooks_cancel_button_xpath}    xpath=//*[@id='edit-webhook-cancel']
${edit_webhooks_save_button_xpath}    xpath=//*[@id='edit-webhook-save']
${project_webhook_edit_id_xpath}    xpath=//*[@id='edit-webhook']
${project_webhook_enable_id_xpath}    xpath=//*[@id='enable-webhook-action']
${project_webhook_disable_id_xpath}    xpath=//*[@id='disable-webhook-action']
${dialog_disable_id_xpath}    xpath=//*[@id='dialog-action-disable']
${dialog_enable_id_xpath}    xpath=//*[@id='dialog-action-enable']
