*** settings ***
Library    JSONLibrary
Resource  ../../resources/Util.robot

*** Keywords ***
#for jsonpath refer to http://goessner.net/articles/JsonPath/ or https://nottyo.github.io/robotframework-jsonlibrary/JSONLibrary.html

${json}=  Load Json From File  testdata.json

Verify User
    @{user}=  Get Value From Json  ${json}  $.users..name
    #verify user exist
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Switch To User Tag
    :For  ${user} In  @{user}
    \    Page Should Contain  ${user}
    Logout Harbor
    #verify user can login
    :For  ${user}  In  @{user}
    \    Sign In Harbor  ${HARBOR_URL}  ${user}  %{HARBOR_PASSWORD}
    \    Logout Harbor

Verify Project
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    :For  ${project}  In  @{project}
    \   Page Should Contain  ${project}

Verify Image Tag
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    :For  ${project}  In  @{projects}
    \    Go Into Project  ${project}
    \    @{repo}=  Get Value From Json  ${json}  $.projects[?(@name=${project})]..repo..name
    \    @{tag}=  Get Value From Json  ${json}  $.projects[?(@name=${project})]..repo..tag
    \    :For  ${repo}  In  @{repo}
    \    \    Go Into Repo  ${repo}
    \    \    :For  ${tag}  In  @{tag}
    \    \    \    Page Should Contain  ${tag}
    \    \    \    Back To Projects

Verify Member Exist
    @{project}=  Get Value From Json  ${json}  $.projects.[*].name
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HAROBR_PASSWORD}
    :For  ${project}  In  @{project}
    \   Go Into Project  ${projet} 
    \   Switch To Member
    \   @{members}=  Get Value From Json  ${json}  $.projects[?(@name=${project})].member..name
    \   :For  ${member}  In  @{members}
    \   \   Page Should Contain  ${member}
    \   Back To Projects

Verify User System Admin Role
    @{user}=  Get Value From Json  ${json}  $.admin..name
    :For  ${user}  in  @{user}
    \    Sign In Harbor  ${HARBOR_URL}  ${user}  %{HARBOR_PASSWORD}
    \    Page Should Contain  Administration 
    \    Logout Harbor
  
Verify System Label
    @{label}=   Get Value From Json  ${json}  $..syslabel..name
    Sign In Harbor  ${HAROBR_URL}  %{HAROBR_ADMIN}  %{HAROBR_PASSWORD}
    Switch To Configure
    Switch To System Labels
    :For  ${label}  In  @{label}
    \   Page Should Contain  ${label}

Verify Project Label
   @{project}= Get Value From Json  ${json}  $.peoject.[*].name
   Sign In Harbor  ${HAROBR_URL}  %{HAROBR_ADMIN}  %{HAROBR_PASSWORD}
   :For  ${project}  In  @{project}
   \    Go Into Project  ${project}
   \    Switch To Project Label
   \    @{projectlabel}=  Get Value From Json  ${json}  $.projects[?(@.name=${project})]..labels..name
   \    :For  ${label}  In  @{projectlabel}
   \    \    Page Should Contain  ${projectlabel}
   \    Back To Projects
      
Verify Endpoint
    @{endpoint}=  Get Value From Json  ${json}  $.endpoint..name
    Sign In Harbor  ${HAROBR_URL}  %{HAROBR_ADMIN}  %{HAROBR_PASSWORD}
    Switch To Registries
    :For  ${endpoint}  In  @{endpoint}
    \    Page Should Contain  ${endpoint}

Verify Replicationrule
    @{replicationrule}=  Get Value From Json  ${json}  $.replicationrule..name
    Sign In Harbor  ${HAROBR_URL}  %{HAROBR_ADMIN}  %{HAROBR_PASSWORD}
    Switch To System Replication
    :For  ${replicationrule}  In  @{replicationrule}
    \    Page Should Contain  ${replicationrule}

Verify Project Setting
    @{projects}=  Get Value From Json  ${json}  $.projects.[*].name
    :For  ${project} In  @{projects}
    \    ${public}=  Get Value From Json  ${json}  $.projects[?(@.name=${projectname})].accesslevel
    \    ${contenttrust}=  Get Value From Json  ${json}  $.projects[?(@.name=${projectname})]..enable_content_trust
    \    ${preventrunning}=  Get Value From Json  ${json}  $.projects[?(@.name=${projectname})]..prevent_vulnerable_images_from_running
    \    ${scanonpush}=  Get Value From Json  ${json}  $.projects[?(@.name=${projectname})]..automatically_scan_images_on_push
    \    Sign In Harbor  ${HAROBR_URL}  %{HAROBR_ADMIN}  %{HAROBR_PASSWORD}
    \    Go Into Project  ${project}
    \    Goto Project Config
    \    Run Keyword If  ${public} == "public"
    \       Checkbox Should Be Checked  //clr-checkbox[@name='public']//label
    \       Else
    \       Checkbox Should Not Be Checked  //clr-checkbox[@name='public']//label
    \    Run Keyword If  ${contenttrust} == "true"
    \       Checkbox Should Be Checked  //clr-checkbox[@name='content-trust']//label
    \       Else
    \       Checkbox Should Not Be Checked  //clr-checkbox[@name='content-trust']//label
    \    Run Keyword If  ${preventrunning} == "true"
    \       Checkbox Should Be Checked  //clr-checkbox[@name='prevent-vulenrability-image']//label
    \       #verify level?page should not contain disabled element
    \       Else 
    \       Checkbox Should Not Be Checked  //clr-checkbox[@name='prevent-vulenrability-image']//label
    \       #Page Should Contain a disabled element
    \    Run Keyword If  ${scanonpush} == "true"
    \       Checkbox Should Be Checked  //clr-checkbox[@name='scan-image-on-push']//label
    \       Else
    \       Checkbox Should Not Be Checked  //clr-checkbox[@name='scan-image-on-push']//label
    \   Back To Projects

Verify System Setting
    ${authtype}=  Get Value From Json  ${json}  $.configuration.authmode
    ${creation}=  Get Value From Json  ${json}  $.configuration..projectcreation
    ${selfreg}=  Get Value From Json  ${json}  $.configuration..selfreg
    ${emailserver}=  Get Value From Json  ${json}  $.configuration..emailserver
    ${emailport}=  Get Value From Json  ${json}  $.configuration..emailport
    ${emailuser}=  Get Value From Json  ${json}  $.configuration..emailuser
    ${emailfrom}=  Get Value From Json  ${json}  $.configuration..emailfrom
    ${token}=  Get Value From Json  ${json}  $.configuration..token
    ${scanschedule}=  Get Value From Json  ${json}  $.configuration..scanall
    Sign In Harbor  ${HARBOR_URL}  %{HARBOR_ADMIN}  %{HARBOR_PASSWORD}
    Switch To Configure
    Page Should Contain  ${authtype}
    Run Keyword If  ${selfreg} == "True"
        Checkbox Should Be Checked  //clr-checkbox[@id='selfReg']//label
        Else
        Checkbox Should Not Be Checked  //clr-checkbox[@id='selfReg']//label
    Page Should Contain  ${creation} 
    Switch To Email
    Page Should Contain  ${emailserver}
    Page Should Contain  ${emailport}
    Page Should Contain  ${emailuser}
    Page Should Contain  ${emailfrom}
    Switch To System Settings
    Page Should Contain  ${token}
    Go To  Vulnerability Config
    Page Should Contain  ${scanschedule}

