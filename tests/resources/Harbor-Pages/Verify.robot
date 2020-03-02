*** settings ***
Resource  ../../resources/Util.robot

*** Keywords ***
#for jsonpath refer to http://goessner.net/articles/JsonPath/ or https://nottyo.github.io/robotframework-jsonlibrary/JSONLibrary.html

Verify User
    [Arguments]    ${json}
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To User Tag
    @{user}=  Get Value From Json  ${json}  $.users..name
    :FOR    ${user}    IN    @{user}
    \    Page Should Contain    ${user}
    Logout Harbor
    #verify user can login
    @{user}=  Get Value From Json  ${json}  $.users..name
    :FOR    ${user}    IN    @{user}
    \    Sign In Harbor    ${HARBOR_URL}  ${user}  ${HARBOR_PASSWORD}
    \    Logout Harbor
    Close Browser

Verify Project
    [Arguments]    ${json}
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    :FOR    ${project}    IN    @{project}
    \    Retry Wait Until Page Contains    ${project}
    Verify Project Metadata  ${json}
    Close Browser

Verify Image Tag
    [Arguments]    ${json}
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    :FOR    ${project}    IN    @{project}
    \    @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
    \    ${has_image}  Set Variable If  @{out_has_image}[0] == ${true}  ${true}  ${false}
    \    Go Into Project  ${project}  has_image=${has_image}
    \    @{repo}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})]..repo..name
    \    Run Keyword If  ${has_image} == ${true}  Loop Image Repo  @{repo}
    \    Navigate To Projects
    Close Browser

Verify Project Metadata
    [Arguments]    ${json}
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    :FOR    ${project}    IN    @{project}
    \    @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
    \    ${has_image}  Set Variable If  @{out_has_image}[0] == ${true}  ${true}  ${false}
    \    Go Into Project  ${project}  has_image=${has_image}
    \    Switch To Project Configuration
    \    Verify Checkbox  ${json}  $.projects[?(@.name=${project})].configuration.public  ${project_config_public_checkbox}
    \    Verify Checkbox  ${json}  $.projects[?(@.name=${project})].configuration.enable_content_trust  ${project_config_content_trust_checkbox}
    \    Verify Checkbox  ${json}  $.projects[?(@.name=${project})].configuration.auto_scan  ${project_config_scan_images_on_push_checkbox}
    \    Verify Checkbox  ${json}  $.projects[?(@.name=${project})].configuration.prevent_vul  ${project_config_prevent_vulnerable_images_from_running_checkbox}
    \    ${ret}    Get Selected List Value    ${project_config_severity_select}
    \    @{severity}=    Get Value From Json    ${json}    $.projects[?(@.name=${project})].configuration.severity
    \    Should Contain    ${ret}    @{severity}[0]
    \    Navigate To Projects
    Close Browser

Verify Checkbox
    [Arguments]    ${json}    ${key}    ${checkbox}
    @{out}=    Get Value From Json    ${json}    ${key}
    Run Keyword If    '@{out}[0]'=='true'    Checkbox Should Be Selected    ${checkbox}
    ...    ELSE    Checkbox Should Not Be Selected    ${checkbox}


Loop Image Repo
    [Arguments]    @{repo}
    :For    ${repo}    In    @{repo}
    \    Page Should Contain  ${repo}

Verify Member Exist
    [Arguments]    ${json}
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    :For    ${project}    In    @{project}
    \   @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
    \   ${has_image}  Set Variable If  @{out_has_image}[0] == ${true}  ${true}  ${false}
    \   Go Into Project  ${project}  has_image=${has_image}
    \   Switch To Member
    \   @{members}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].member..name
    \   Loop Member  @{members}
    \   Navigate To Projects
    Close Browser

Loop Member
    [Arguments]    @{members}
    :For    ${member}    In    @{members}
    \    Page Should Contain    ${member}

Verify User System Admin Role
    [Arguments]    ${json}
    @{user}=  Get Value From Json  ${json}  $.admin..name
    Init Chrome Driver
    :FOR    ${user}    IN    @{user}
    \    Sign In Harbor  ${HARBOR_URL}  ${user}  ${HARBOR_PASSWORD}
    \    Page Should Contain  Administration
    \    Logout Harbor
    Close Browser

Verify System Label
    [Arguments]    ${json}
    @{label}=   Get Value From Json  ${json}  $..syslabel..name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Switch To System Labels
    :For    ${label}    In    @{label}
    \   Page Should Contain    ${label}
    Close Browser

Verify Project Label
   [Arguments]    ${json}
   @{project}= Get Value From Json  ${json}  $.peoject.[*].name
   Init Chrome Driver
   Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    :For    ${project}    In    @{project}
    \    @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
    \    ${has_image}  Set Variable If  @{out_has_image}[0] == ${true}  ${true}  ${false}
    \    Go Into Project  ${project}  has_image=${has_image}
    \    Switch To Project Label
    \    @{projectlabel}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})]..labels..name
    \    :For    ${label}    In    @{label}
    \    \    Page Should Contain    ${projectlabel}
    \    Navigate To Projects
   Close Browser

Verify Endpoint
    [Arguments]    ${json}
    @{endpoint}=  Get Value From Json  ${json}  $.endpoint..name
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Registries
    :For    ${endpoint}    In    @{endpoint}
    \    Page Should Contain    ${endpoint}
    Close Browser

Verify Replicationrule
    [Arguments]    ${json}
    @{replicationrules}=    Get Value From Json    ${json}    $.replicationrule.[*].rulename
    @{endpoints}=    Get Value From Json    ${json}    $.endpoint.[*].name
    : FOR    ${replicationrule}    IN    @{replicationrules}
    \    Init Chrome Driver
    \    Log To Console    -----replicationrule-----"${replicationrule}"------------
    \    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    \    Switch To Replication Manage
    \    Select Rule And Click Edit Button    ${replicationrule}
    \    @{is_src_registry}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].is_src_registry
    \    @{trigger_type}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].trigger_type
    \    @{name_filters}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].name_filters
    \    @{tag_filters}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].tag_filters
    \    @{dest_namespace}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].dest_namespace
    \    @{cron}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].cron
    \    @{is_src_registry}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].is_src_registry
    \    Log To Console    -----is_src_registry-----@{is_src_registry}[0]------------
    \    @{endpoint}=    Get Value From Json    ${json}    $.replicationrule[?(@.rulename=${replicationrule})].endpoint
    \    Log To Console    -----endpoint-----@{endpoint}------------
    \    ${endpoint0}=   Set Variable    @{endpoint}[0]
    \    Log To Console    -----endpoint0-----${endpoint0}------------
    \    @{endpoint_type}=    Get Value From Json    ${json}    $.endpoint[?(@.name=${endpoint0})].type
    \    Retry Textfield Value Should Be    ${source_project}    @{name_filters}[0]
    \    Retry Textfield Value Should Be    ${filter_tag}    @{tag_filters}[0]
    \    Retry Textfield Value Should Be    ${rule_name_input}    ${replicationrule}
    \    Retry Textfield Value Should Be    ${dest_namespace_xpath}    @{dest_namespace}[0]
    \    Log To Console    -----endpoint_type-----@{endpoint_type}[0]------------
    \    ${registry}=    Set Variable If    "@{endpoint_type}[0]"=="harbor"    ${endpoint0}-https://${IP}    ${endpoint0}-https://hub.docker.com
    \    Log To Console    -------registry---${registry}------------
    \    Run Keyword If    '@{is_src_registry}[0]' == '${true}'    Retry List Selection Should Be    ${src_registry_dropdown_list}    ${registry}
    \    ...    ELSE    Retry List Selection Should Be    ${dest_registry_dropdown_list}    ${registry}
    \    #\    Retry List Selection Should Be    ${rule_resource_selector}    ${resource_type}
    \    Retry List Selection Should Be    ${rule_trigger_select}    @{trigger_type}[0]
    \    Run Keyword If    '@{trigger_type}[0]' == 'scheduled'    Log To Console    ----------@{trigger_type}[0]------------
    \    Run Keyword If    '@{trigger_type}[0]' == 'scheduled'    Retry Textfield Value Should Be    ${targetCron_id}    @{cron}[0]
    Close Browser

Verify Project Setting
    [Arguments]    ${json}
    @{projects}=  Get Value From Json  ${json}  $.projects.[*].name
    :For    ${project}    In    @{Projects}
    \    ${public}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].accesslevel
    \    ${contenttrust}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})]..enable_content_trust
    \    ${preventrunning}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})]..prevent_vulnerable_images_from_running
    \    ${scanonpush}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})]..automatically_scan_images_on_push
    \    Init Chrome Driver
    \    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    \    @{out_has_image}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})].has_image
    \    ${has_image}  Set Variable If  @{out_has_image}[0] == ${true}  ${true}  ${false}
    \    Go Into Project  ${project}  has_image=${has_image}
    \    Goto Project Config
    \    Run Keyword If  ${public} == "public"  Checkbox Should Be Checked  //clr-checkbox-wrapper[@name='public']//label
    \    Run Keyword If  ${contenttrust} == "true"  Checkbox Should Be Checked  //clr-checkbox-wrapper[@name='content-trust']//label
    \    Run Keyword If  ${contenttrust} == "false"  Checkbox Should Not Be Checked  //clr-checkbox-wrapper[@name='content-trust']//label
    \    Run Keyword If  ${preventrunning} == "true"  Checkbox Should Be Checked  //*[@id='prevent-vulenrability-image']//clr-checkbox-wrapper//label
    \    Run Keyword If  ${preventrunning} == "false"  Checkbox Should Not Be Checked    //*[@id='prevent-vulenrability-image']//clr-checkbox-wrapper//label
    \    Run Keyword If  ${scanonpush} == "true"  Checkbox Should Be Checked  //clr-checkbox-wrapper[@id='scan-image-on-push-wrapper']//input
    \    Run Keyword If  ${scanonpush} == "true"  Checkbox Should Not Be Checked  //clr-checkbox-wrapper[@id='scan-image-on-push-wrapper']//input
    \   Close Browser

Verify System Setting
    [Arguments]    ${json}
    @{authtype}=  Get Value From Json  ${json}  $.configuration.authmode
    @{creation}=  Get Value From Json  ${json}  $.configuration..projectcreation
    @{selfreg}=  Get Value From Json  ${json}  $.configuration..selfreg
    @{emailserver}=  Get Value From Json  ${json}  $.configuration..emailserver
    @{emailport}=  Get Value From Json  ${json}  $.configuration..emailport
    @{emailuser}=  Get Value From Json  ${json}  $.configuration..emailuser
    @{emailfrom}=  Get Value From Json  ${json}  $.configuration..emailfrom
    @{token}=  Get Value From Json  ${json}  $.configuration..token
    @{scanschedule}=  Get Value From Json  ${json}  $.configuration..scanall
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Page Should Contain  @{authtype}[0]
    Run Keyword If  @{selfreg}[0] == 'True'  Checkbox Should Be Checked  //clr-checkbox-wrapper[@id='selfReg']//label
    Run Keyword If  @{selfreg}[0] == 'False'  Checkbox Should Not Be Checked  //clr-checkbox-wrapper[@id='selfReg']//label
    Switch To Email
    Textfield Value Should Be  xpath=//*[@id='mailServer']  @{emailserver}[0]
    Textfield Value Should Be  xpath=//*[@id='emailPort']  @{emailport}[0]
    Textfield Value Should Be  xpath=//*[@id='emailUsername']  @{emailuser}[0]
    Textfield Value Should Be  xpath=//*[@id='emailFrom']  @{emailfrom}[0]
    Switch To System Settings
    ${ret}  Get Selected List Value  xpath=//select[@id='proCreation']
    Should Be Equal As Strings  ${ret}  @{creation}[0]
    Token Must Be Match  @{token}[0]
    #ToDo:These 2 lines below should be uncommented right after issue 9211 was fixed
    #Switch To Vulnerability Page
    #Page Should Contain  None
    Close Browser

