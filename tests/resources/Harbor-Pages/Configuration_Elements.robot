# Copyright Project Harbor Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License

*** Settings ***
Documentation  This resource provides any keywords related to the Harbor private registry appliance

*** Variables ***
${project_create_xpath}  //clr-dg-action-bar//button[contains(.,'New')]
${self_reg_xpath}  //input[@id='selfReg']
${test_ldap_xpath}  //*[@id='ping-test']
${config_save_button_xpath}  //config//div/button[contains(.,'SAVE')]
${config_auth_save_button_xpath}  //*[@id='config_auth_save']
${config_system_save_button_xpath}  //*[@id='config_system_save']
${config_security_save_button_xpath}  //*[@id='security_save']
${vulnerbility_save_button_xpath}  //*[@id='config-save']
${configuration_xpath}  //clr-main-container//clr-vertical-nav//a[contains(.,' Configuration ')]
${configuration_system_tabsheet_id}  //*[@id='config-system']
${configuration_security_tabsheet_id}  //*[@id='config-security']
${configuration_authentication_tabsheet_id}  //*[@id="config-auth"]
${configuration_project_quotas_tabsheet_id}  //*[@id='config-quotas']
${configuration_system_wl_add_btn}    //*[@id='show-add-modal-button']
${configuration_system_wl_textarea}    //*[@id='allowlist-textarea']
${configuration_system_wl_add_confirm_btn}    //*[@id='add-to-system']
${configuration_system_wl_delete_a_cve_id_icon}    //app-security//form/section//ul/li[1]/a[2]/clr-icon
${configuration_sys_repo_readonly_chb_id}  //*[@id='repo_read_only_lbl']
${cve_allowlist_expires_btn}  //clr-date-container[.//div[@class='clr-input-group' and not(@hidden)]]//button
${cve_allowlist_expires_yesterday}  //td[.//button[@class='day-btn is-today']]/preceding::td[1]
${cve_allowlist_expires_tomorrow}  //td[.//button[@class='day-btn is-today']]/following::td[1]
${cfg_auth_automatic_onboarding_checkbox}  //clr-checkbox-wrapper//label[contains(@class,'clr-control-label') and contains(@for,'oidcAutoOnboard')]
${cfg_auth_user_name_claim_input}  //*[@id='oidcUserClaim']

${cfg_auth_ldap_group_admin_dn}  //*[@id='ldapGroupAdminDN']


${distribution_add_btn_id}  //*[@id='new-instance']
${distribution_provider_select_id}  //*[@id='provider']
${distribution_provider_authmode_id}  //clr-main-container//clr-radio-container//label[contains(@class,'clr-control-label') and contains(.,'OAuth')]
${distribution_provider_authcode_id}  //*[@id='auth_data_token']
${distribution_name_input_id}  //*[@id='name']
${distribution_endpoint_id}  //*[@id='endpoint']
${distribution_description_id}  //*[@id='description']
${distribution_auth_none_mode_ratio_id}  //*[@id='none_mode']
${distribution_auth_basic_mode_ratio_id}  //*[@id='basic_mode']
${distribution_auth_oauth_mode_ratio_id}  //*[@id='token_mode']
${distribution_enable_checkbox_id}  //*[@id='enabled']
${distribution_insecure_checkbox_id}  //*[@id='insecure']
${distribution_add_save_btn_id}  //*[@id='instance-ok']
${distribution_action_btn_id}  //*[@id='member-action']
${distribution_del_btn_id}  //*[@id='distribution-delete']
${distribution_edit_btn_id}  //*[@id='distribution-edit']
${filter_dist_btn}  //hbr-filter//clr-icon[contains(@class,'search-btn')]
${filter_dist_input}  //hbr-filter//input

${audit_log_forward_syslog_endpoint_input_id}  //*[@id='auditLogForwardEndpoint']
${skip_audit_log_database_checkbox}  //*[@id='skipAuditLogDatabase']
${skip_audit_log_database_label}  //clr-checkbox-wrapper//label[contains(@class,'clr-control-label') and contains(@for,'skipAuditLogDatabase')]
${retain_image_last_pull_time_label}  //clr-checkbox-wrapper//label[contains(@class,'clr-control-label') and contains(@for,'scannerSkipUpdatePullTime')]
${retain_image_last_pull_time_checkbox}  //*[@id='scannerSkipUpdatePullTime']
${banner_message_input_id}  //*[@id='banner-message']
${banner_message_type_select_id}  //*[@id='banner-message-type']
${banner_message_closable_checkbox}  //div[contains(@class,'clr-checkbox-wrapper')]//label[contains(@class,'clr-control-label')]
${banner_message_from_date}  (//clr-date-container)[1]//button
${banner_message_to_date}  (//clr-date-container)[2]//button
${banner_message_date_next_month}  //button[@aria-label='Next month']
${banner_message_close_alert}  //clr-alert//button[@aria-label='Close alert']
${banner_message_alert}  //app-app-level-alerts
