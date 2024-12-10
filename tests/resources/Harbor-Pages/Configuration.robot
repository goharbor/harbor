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
Resource  ../../resources/Util.robot

*** Variables ***

*** Keywords ***
Init LDAP
    ${rc}  ${output}=  Run And Return Rc And Output  ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'
    Log  ${output}
    Sleep  2
    Input Text  xpath=//*[@id='ldapUrl']  ldaps://${output}
    Sleep  1
    Input Text  xpath=//*[@id='ldapSearchDN']  cn=admin,dc=example,dc=com
    Sleep  1
    Input Text  xpath=//*[@id='ldapSearchPwd']  admin
    Sleep  1
    Input Text  xpath=//*[@id='ldapBaseDN']  dc=example,dc=com
    Sleep  1
    Input Text  xpath=//*[@id='ldapFilter']  (&(objectclass=inetorgperson)(memberof=cn=harbor_users,ou=groups,dc=example,dc=com))
    Sleep  1
    Input Text  xpath=//*[@id='ldapUid']  cn
    Sleep  1
    Disable Ldap Verify Cert Checkbox
    Retry Element Click  xpath=${config_auth_save_button_xpath}
    Sleep  2
    Retry Element Click  xpath=/html/body/harbor-app/harbor-shell/clr-main-container/div/div/config/div/div/div/button[3]
    Sleep  1

Switch To Configure
    Retry Element Click  xpath=${configuration_xpath}

Test Ldap Connection
    ${rc}  ${output}=  Run And Return Rc And Output  ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'
    Log  ${output}
    Sleep  2
    Input Text  xpath=//*[@id='ldapUrl']  ldaps://${output}
    Sleep  1
    Input Text  xpath=//*[@id='ldapSearchDN']  cn=admin,dc=example,dc=com
    Sleep  1
    Input Text  xpath=//*[@id='ldapSearchPwd']  admin
    Sleep  1
    Input Text  xpath=//*[@id='ldapBaseDN']  dc=example,dc=com
    Sleep  1
    Input Text  xpath=//*[@id='ldapUid']  cn
    Sleep  1

    # default is checked, click test connection to verify fail as no cert.
    Retry Element Click  xpath=${test_ldap_xpath}
    Sleep  1
    Wait Until Page Contains  Failed to verify LDAP server with error
    Sleep  5

    Disable Ldap Verify Cert Checkbox
    # ldap checkbox unchecked, click test connection to verify success.
    Sleep  1
    Retry Element Click  xpath=${test_ldap_xpath}
    Wait Until Page Contains  Connection to LDAP server is verified  timeout=15

Test LDAP Server Success
    Retry Element Click  xpath=${test_ldap_xpath}
    Wait Until Page Contains  Connection to LDAP server is verified  timeout=15

Disable Ldap Verify Cert Checkbox
    Mouse Down  xpath=//*[@id='clr-checkbox-ldapVerifyCert']
    Mouse Up  xpath=//*[@id='clr-checkbox-ldapVerifyCert']
    Sleep  2
    Ldap Verify Cert Checkbox Should Be Disabled

Ldap Verify Cert Checkbox Should Be Disabled
    Checkbox Should Not Be Selected  xpath=//*[@id='clr-checkbox-ldapVerifyCert']

Set Pro Create Admin Only
    #set limit to admin only
    Retry Element Click  xpath=${configuration_xpath}
    Sleep  2
    Retry Element Click  xpath=${configuration_system_tabsheet_id}
    Sleep  1
    Retry Element Click  xpath=//select[@id='proCreation']
    Retry Element Click  xpath=//select[@id='proCreation']//option[@value='adminonly']
    Sleep  1
    Retry Element Click  xpath=${config_system_save_button_xpath}

Set Pro Create Every One
    Retry Element Click  xpath=${configuration_xpath}
    sleep  1
    #set limit to Every One
    Retry Element Click  xpath=${configuration_system_tabsheet_id}
    Sleep  1
    Retry Element Click  xpath=//select[@id='proCreation']
    Retry Element Click  xpath=//select[@id='proCreation']//option[@value='everyone']
    Sleep  1
    Retry Element Click  xpath=${config_system_save_button_xpath}
    Sleep  2

Disable Self Reg
    Retry Element Click  xpath=${configuration_xpath}
    Mouse Down  xpath=${self_reg_xpath}
    Mouse Up  xpath=${self_reg_xpath}
    Sleep  1
    Self Reg Should Be Disabled
    Retry Element Click  xpath=${config_auth_save_button_xpath}
    Sleep  1

Enable Self Reg
    Mouse Down  xpath=${self_reg_xpath}
    Mouse Up  xpath=${self_reg_xpath}
    Sleep  1
    Self Reg Should Be Enabled
    Retry Element Click  xpath=${config_auth_save_button_xpath}
    Sleep  1

Self Reg Should Be Disabled
    Checkbox Should Not Be Selected  xpath=${self_reg_xpath}

Self Reg Should Be Enabled
    Checkbox Should Be Selected  xpath=${self_reg_xpath}

Project Creation Should Display
    Retry Wait Until Page Contains Element  xpath=${project_create_xpath}

Project Creation Should Not Display
    Retry Wait Until Page Not Contains Element  xpath=${project_create_xpath}

Switch To System Settings
    Retry Element Click  xpath=${configuration_xpath}
    Retry Element Click  xpath=${configuration_system_tabsheet_id}

Switch To Project Quotas
    Sleep  1
    Retry Element Click  xpath=${configuration_xpath}
    Sleep  1
    Retry Element Click  xpath=//clr-main-container//clr-vertical-nav//a[contains(.,'Project Quotas')]
    Sleep  1

Switch To Distribution
    Retry Element Click  xpath=//clr-main-container//clr-vertical-nav-group//span[contains(.,'Distributions')]

Switch To Robot Account
    Sleep  1
    Retry Element Click  xpath=//clr-main-container//clr-vertical-nav-group//span[contains(.,'Robot Accounts')]
    Sleep  1

Modify Token Expiration
    [Arguments]  ${minutes}
    Input Text  xpath=//*[@id='tokenExpiration']  ${minutes}
    Click Button  xpath=${config_system_save_button_xpath}
    Sleep  1

Token Must Be Match
    [Arguments]  ${minutes}
    Textfield Value Should Be  xpath=//*[@id='tokenExpiration']  ${minutes}

Robot Account Token Must Be Match
    [Arguments]  ${days}
    Textfield Value Should Be  xpath=//*[@id='robotTokenExpiration']  ${days}

## Replication
Check Verify Remote Cert
    Mouse Down  xpath=//*[@id='clr-checkbox-verifyRemoteCert']
    Mouse Up  xpath=//*[@id='clr-checkbox-verifyRemoteCert']
    Retry Element Click  xpath=${config_save_button_xpath}
    Sleep  1

Switch To System Replication
    Sleep  1
    Switch To Configure
    Retry Element Click  xpath=//*[@id='config-replication']
    Sleep  1

Should Verify Remote Cert Be Enabled
    Checkbox Should Not Be Selected  xpath=//*[@id='clr-checkbox-verifyRemoteCert']

Set Scan All To None
    Retry Element Click  //vulnerability-config//select
    Retry Element Click  //vulnerability-config//select/option[@value='none']
    sleep  1
    Retry Element Click  ${vulnerbility_save_button_xpath}

Set Scan All To Daily
    Retry Element Click  //vulnerability-config//select
    Retry Element Click  //vulnerability-config//select/option[@value='daily']
    sleep  1
    Retry Element Click  ${vulnerbility_save_button_xpath}

Click Scan Now
    Retry Element Click  //vulnerability-config//button[contains(.,'SCAN')]


Enable Read Only
    ${rc}  ${output}=  Run And Return Rc And Output  curl -u admin:Harbor12345 -s --insecure -H "Content-Type: application/json" -X PUT -d '{"read_only":true}' "https://${ip}/api/v2.0/configurations"
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0

Disable Read Only
    ${rc}  ${output}=  Run And Return Rc And Output  curl -u admin:Harbor12345 -s --insecure -H "Content-Type: application/json" -X PUT -d '{"read_only":false}' "https://${ip}/api/v2.0/configurations"
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0

## System labels
Switch To System Labels
    Sleep  1
    Retry Element Click  xpath=//clr-main-container//clr-vertical-nav//a[contains(.,'Labels')]

Switch To Configuration System Setting
    Sleep  1
    Retry Element Click  xpath=${configuration_xpath}
    Retry Element Click  xpath=${configuration_system_tabsheet_id}

Switch To Configuration Security
    Retry Element Click  xpath=${configuration_xpath}
    Retry Double Keywords When Error  Retry Element Click  xpath=${configuration_security_tabsheet_id}  Retry Wait Until Page Contains  Deployment security

Switch To Configuration Project Quotas
    Sleep  1
    Retry Element Click  xpath=//clr-main-container//clr-vertical-nav//a[contains(.,'Project Quotas')]

Create New Labels
    [Arguments]  ${labelname}
    Retry Element Click  xpath=//button[contains(.,'New Label')]
    Sleep  1
    Input Text  xpath=//*[@id='name']  ${labelname}
    Sleep  1
    Retry Element Click  xpath=//hbr-create-edit-label//clr-dropdown/clr-icon
    Sleep  1
    Retry Element Click  xpath=//hbr-create-edit-label//clr-dropdown-menu/label[1]
    Sleep  1
    Input Text  xpath=//*[@id='description']  global
    Retry Element Click  xpath=//div/form/section/label[4]/button[2]
    Wait Until Page Contains  ${labelname}

Update A Label
    [Arguments]  ${labelname}
    Retry Element Click  xpath=//clr-dg-row[contains(.,'${labelname}')]//div[contains(@class,'clr-checkbox-wrapper')]//label[contains(@class,'clr-control-label')]
    Sleep  1
    Retry Element Click  xpath=//button[contains(.,'Edit')]
    Sleep  1
    Input Text  xpath=//*[@id='name']  ${labelname}1
    Sleep  1
    Retry Element Click  xpath=//hbr-create-edit-label//form/section//button[2]
    Wait Until Page Contains  ${labelname}1

Delete A Label
    [Arguments]  ${labelname}
    Retry Element Click  xpath=//clr-dg-row[contains(.,'${labelname}')]//div[contains(@class,'clr-checkbox-wrapper')]//label[contains(@class,'clr-control-label')]
    Sleep  1
    Retry Element Click  xpath=//button[contains(.,'Delete')]
    Sleep  3
    Retry Element Click  xpath=//clr-modal//div//button[contains(.,'DELETE')]
    Wait Until Page Contains Element  //*[@id='contentAll']//div[contains(.,'${labelname}')]/../div/clr-icon[@shape='success-standard']

Add Items To System CVE Allowlist
    [Arguments]    ${cve_id}
    Retry Element Click    ${configuration_system_wl_add_btn}
    Retry Text Input    ${configuration_system_wl_textarea}    ${cve_id}
    Retry Element Click    ${configuration_system_wl_add_confirm_btn}
    Retry Element Click    ${config_security_save_button_xpath}

Delete Top Item In System CVE Allowlist
    [Arguments]  ${count}=1
    FOR  ${idx}  IN RANGE  1  ${count}
        Retry Element Click  ${configuration_system_wl_delete_a_cve_id_icon}
    END
    Retry Element Click  ${config_security_save_button_xpath}

Set CVE Allowlist Expires
    [Arguments]  ${expired}
    Retry Button Click  ${cve_allowlist_expires_btn}
    ${element}=  Set Variable If  ${expired}  ${cve_allowlist_expires_yesterday}  ${cve_allowlist_expires_tomorrow}
    Retry Element Click  ${element}
    Retry Element Click  //button[contains(.,'SAVE')]

Get Project Count Quota Text From Project Quotas List
    [Arguments]    ${project_name}
    Switch To Project Quotas
    ${count_quota}=    Get Text    xpath=//project-quotas//clr-datagrid//clr-dg-row[contains(.,'${project_name}')]//clr-dg-cell[3]//label
    [Return]  ${count_quota}

Get Project Storage Quota Text From Project Quotas List
    [Arguments]    ${project_name}
    Switch To Configure
    Switch To Project Quotas
    ${storage_quota}=    Get Text    xpath=//project-quotas//clr-datagrid//clr-dg-row[contains(.,'${project_name}')]//clr-dg-cell[3]//label
    [Return]  ${storage_quota}

Check Automatic Onboarding And Save
    Switch To Configure
    Retry Element Click  ${cfg_auth_automatic_onboarding_checkbox}
    Retry Element Click  xpath=${config_auth_save_button_xpath}

Set User Name Claim And Save
    [Arguments]    ${type}
    Switch To Configure
    Retry Clear Element Text  ${cfg_auth_user_name_claim_input}
    Run Keyword If  '${type}'=='${null}'  Retry Text Input  ${cfg_auth_user_name_claim_input}  anytext
    ...  ELSE  Retry Text Input  ${cfg_auth_user_name_claim_input}  ${type}
    Retry Element Click  xpath=${config_auth_save_button_xpath}

Select Distribution
    [Arguments]    ${name}
    Retry Element Click    //clr-dg-row[contains(.,'${name}')]//div[contains(@class,'clr-checkbox-wrapper')]/label[contains(@class,'clr-control-label')]

Distribution Exist
    [Arguments]  ${name}  ${endpoint}
    Retry Wait Until Page Contains Element  //clr-dg-row[contains(.,'${name}') and contains(.,'${endpoint}')]

Distribution Not Exist
    [Arguments]  ${name}  ${endpoint}
    Retry Wait Until Page Not Contains Element  //clr-dg-row[contains(.,'${name}') and contains(.,'${endpoint}')]

Filter Distribution List
    [Arguments]  ${name}  ${endpoint}  ${exsit}=${true}
    Retry Double Keywords When Error  Retry Element Click  ${filter_dist_btn}  Wait Until Element Is Visible And Enabled  ${filter_dist_input}
    Retry Text Input  ${filter_dist_input}  ${name}
    Run Keyword If  ${exsit}==${true}    Distribution Exist  ${name}  ${endpoint}
    ...  ELSE  Distribution Not Exist  ${name}  ${endpoint}

Select Provider
    [Arguments]    ${provider}
    Retry Element Click    ${distribution_provider_select_id}
    Retry Element Click    ${distribution_provider_select_id}//option[contains(.,'${provider}')]

Set Authcode
    [Arguments]    ${auth_code}
    Retry Element Click    ${distribution_provider_authmode_id}
    Retry Text Input  ${distribution_provider_authcode_id}  ${auth_code}

Create An New Distribution
    [Arguments]    ${provider}  ${name}  ${endpoint}  ${auth_token}
    Switch To Distribution
    Retry Element Click  ${distribution_add_btn_id}
    Select Provider  ${provider}
    Retry Text Input  ${distribution_name_input_id}  ${name}
    Retry Text Input  ${distribution_endpoint_id}  ${endpoint}
    Set Authcode  ${auth_token}
    Retry Double Keywords When Error  Retry Element Click  ${distribution_add_save_btn_id}  Retry Wait Until Page Not Contains Element  xpath=${distribution_add_save_btn_id}
    Distribution Exist  ${name}  ${endpoint}

Delete A Distribution
    [Arguments]    ${name}  ${endpoint}  ${deletable}=${true}
    ${is_exsit}    evaluate    not ${deletable}
    Switch To Distribution
    Filter Distribution List  ${name}  ${endpoint}
    Retry Double Keywords When Error  Select Distribution   ${name}  Wait Until Element Is Visible  //clr-datagrid//clr-dg-footer//clr-checkbox-wrapper/label[contains(@class,'clr-control-label')]
    Retry Double Keywords When Error  Retry Element Click  ${distribution_action_btn_id}  Wait Until Element Is Visible And Enabled  ${distribution_del_btn_id}
    Retry Double Keywords When Error  Retry Element Click  ${distribution_del_btn_id}  Wait Until Element Is Visible And Enabled  ${delete_confirm_btn}
    Retry Double Keywords When Error  Retry Element Click  ${delete_confirm_btn}  Retry Wait Until Page Not Contains Element  ${delete_confirm_btn}
    Filter Distribution List  ${name}  ${endpoint}  exsit=${is_exsit}

Edit A Distribution
    [Arguments]    ${name}  ${endpoint}  ${new_endpoint}=${null}
    Switch To Distribution
    Filter Distribution List  ${name}  ${endpoint}
    Retry Double Keywords When Error  Select Distribution   ${name}  Wait Until Element Is Visible  //clr-datagrid//clr-dg-footer//clr-checkbox-wrapper/label[contains(@class,'clr-control-label')]  times=9
    Retry Double Keywords When Error  Retry Element Click  ${distribution_action_btn_id}  Wait Until Element Is Visible And Enabled  ${distribution_edit_btn_id}
    Retry Double Keywords When Error  Retry Element Click  ${distribution_edit_btn_id}  Wait Until Element Is Visible And Enabled  ${distribution_name_input_id}
    Retry Text Input  ${distribution_endpoint_id}  ${new_endpoint}
    Retry Double Keywords When Error  Retry Element Click  ${distribution_add_save_btn_id}  Retry Wait Until Page Not Contains Element  xpath=${distribution_add_save_btn_id}
    Filter Distribution List  ${name}  ${new_endpoint}
    Distribution Exist  ${name}  ${new_endpoint}

Set Audit Log Forward
    [Arguments]  ${syslog_endpoint}  ${expected_msg}
    Switch To System Settings
    Run Keyword If  '${syslog_endpoint}' == '${null}'  Press Keys  ${audit_log_forward_syslog_endpoint_input_id}  CTRL+a  BACKSPACE
    ...  ELSE  Retry Text Input  ${audit_log_forward_syslog_endpoint_input_id}  ${syslog_endpoint}
    Retry Double Keywords When Error  Retry Element Click  ${config_save_button_xpath}  Retry Wait Until Page Contains  ${expected_msg}

Enable Skip Audit Log Database
    Switch To System Settings
    Retry Double Keywords When Error  Click Element  ${skip_audit_log_database_label}  Checkbox Should Be Selected  ${skip_audit_log_database_checkbox}
    Retry Double Keywords When Error  Retry Element Click  ${config_save_button_xpath}  Retry Wait Until Page Contains  Configuration has been successfully saved.

Set Up Retain Image Last Pull Time
    [Arguments]  ${action}
    Run Keyword If  '${action}'=='enable'  Retry Double Keywords When Error  Click Element  ${retain_image_last_pull_time_label}  Checkbox Should Be Selected  ${retain_image_last_pull_time_checkbox}
    ...  ELSE  Retry Double Keywords When Error  Click Element  ${retain_image_last_pull_time_label}  Checkbox Should Not Be Selected  ${retain_image_last_pull_time_checkbox}
    Retry Double Keywords When Error  Retry Element Click  ${config_save_button_xpath}  Retry Wait Until Page Contains  Configuration has been successfully saved.

Set Banner Message
    [Arguments]  ${message}  ${message_type}=${null}  ${closable}=${null}  ${in_duration}=${null}
    IF  '${message}' != '${null}'
        Retry Text Input  ${banner_message_input_id}  ${message}
        Select From List By Value  ${banner_message_type_select_id}  ${message_type}
        Run Keyword If  '${closable}' == '${true}'  Retry Element Click  ${banner_message_closable_checkbox}
        IF  '${in_duration}' == '${true}'
            Retry Element Click  ${banner_message_from_date}
            Retry Element Click  //td[.//button[contains(@class,'is-today')]]
            Retry Element Click  ${banner_message_to_date}
            Retry Element Click  ${banner_message_date_next_month}
            Retry Element Click  (//td[.//button[text()=' 1 ']])[1]
        ELSE IF  '${in_duration}' == '${false}'
            Retry Element Click  ${banner_message_from_date}
            Retry Element Click  ${banner_message_date_next_month}
            Retry Element Click  (//td[.//button[text()=' 1 ']])[1]
            Retry Element Click  ${banner_message_to_date}
            Retry Element Click  ${banner_message_date_next_month}
            Retry Element Click  (//td[.//button[text()=' 1 ']])[1]
        END
    ELSE
        Clear Field Of Characters  ${banner_message_input_id}  23
    END
    Retry Double Keywords When Error  Retry Element Click  ${config_save_button_xpath}  Retry Wait Until Page Contains  Configuration has been successfully saved.

Check Banner Message
    [Arguments]  ${message}  ${message_type}=${null}  ${closable}=${null}
    IF  '${message}' == '${null}'
        Retry Wait Element Not Visible  ${banner_message_alert}
    ELSE
        Retry Wait Element  //span[text()='${message}']
        ${message_type_class}=  Create Dictionary  success=alert-success  info=alert-info  warning=alert-warning  danger=alert-danger
        Retry Wait Element  //app-app-level-alerts//clr-alerts[contains(@class,'${message_type_class}[${message_type}]')]
        Run Keyword If  '${closable}' == '${true}'  Retry Wait Element  ${banner_message_close_alert}
    END

Check Banner Message on other pages
    [Arguments]  ${message}  ${message_type}=${null}  ${closable}=${null}
    @{pages}=	 Create List  harbor/projects  harbor/logs  harbor/users  harbor/robot-accounts  harbor/Registries
    ...  harbor/replications  harbor/distribution/instances  harbor/labels  harbor/project-quotas  harbor/interrogation-services/scanners
    ...  harbor/interrogation-services/vulnerability  harbor/interrogation-services/security-hub  harbor/clearing-job/gc
    ...  harbor/clearing-job/audit-log-purge  harbor/job-service-dashboard/pending-jobs  harbor/job-service-dashboard/schedules
    ...  harbor/job-service-dashboard/workers  harbor/configs/auth  harbor/configs/security  harbor/configs/setting
    ...  harbor/projects/1/repositories  harbor/projects/1/members  harbor/projects/1/labels  harbor/projects/1/scanner
    ...  harbor/projects/1/p2p-provider/policies  harbor/projects/1/tag-strategy/tag-retention  harbor/projects/1/tag-strategy/immutable-tag
    ...  harbor/projects/1/robot-account  harbor/projects/1/webhook  harbor/projects/1/logs  harbor/projects/1/configs
    FOR  ${page}  IN  @{pages}
        Go To  ${HARBOR_URL}/${page}
        Check Banner Message  ${message}  ${message_type}  ${closable}
    END
    Logout Harbor
    Check Banner Message  ${message}  ${message_type}  ${closable}
