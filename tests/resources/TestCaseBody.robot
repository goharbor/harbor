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
Documentation  This resource wrap test case body

*** Variables ***

*** Keywords ***
Body Of Manage project publicity
    Init Chrome Driver
    ${d}=    Get Current Date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Create An New Project  project${d}  public=true

    Push image  ${ip}  user007  Test1@34  project${d}  hello-world:latest
    Pull image  ${ip}  user008  Test1@34  project${d}  hello-world:latest

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Display  project${d}
    Search Private Projects
    Project Should Not Display  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Make Project Private  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Not Display  project${d}
    Cannot Pull image  ${ip}  user008  Test1@34  project${d}  hello-world:latest

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Make Project Public  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Display  project${d}
    Close Browser

Body Of Scan A Tag In The Repo
    [Arguments]  ${image_argument}  ${tag_argument}
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user023  Test1@34
    Create An New Project  project${d}
    Go Into Project  project${d}  has_image=${false}
    Push Image  ${ip}  user023  Test1@34  project${d}  ${image_argument}:${tag_argument}
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image_argument}
    Scan Repo  ${tag_argument}  Succeed
    Summary Chart Should Display  ${tag_argument}
    Pull Image  ${ip}  user023  Test1@34  project${d}  ${image_argument}  ${tag_argument}
    # Edit Repo Info
    Close Browser

Body Of List Helm Charts
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user027  Test1@34
    Create An New Project  project${d}
    Go Into Project  project${d}  has_image=${false}

    Switch To Project Charts
    Upload Chart files
    Go Into Chart Version  ${prometheus_chart_name}
    Retry Wait Until Page Contains  ${prometheus_chart_version}
    Go Into Chart Detail  ${prometheus_chart_version}

    # Summary tab
    Retry Wait Until Page Contains Element  ${summary_markdown}
    Retry Wait Until Page Contains Element  ${summary_container}

    # Dependency tab
    Retry Double Keywords When Error  Retry Element Click  xpath=${detail_dependency}  Retry Wait Until Page Contains Element  ${dependency_content}

    # Values tab
    Retry Double Keywords When Error  Retry Element Click  xpath=${detail_value}  Retry Wait Until Page Contains Element  ${value_content}

    Go Into Project  project${d}  has_image=${false}
    Switch To Project Charts
    Multi-delete Chart Files  ${prometheus_chart_name}  ${harbor_chart_name}
    Close Browser

Body Of Admin Push Signed Image
    [Arguments]  ${image}=tomcat  ${with_remove}=${false}
    Enable Notary Client

    Docker Pull  ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}
    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group0-Util/notary-push-image.sh ${ip} library ${image} latest ${notaryServerEndpoint} ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}:latest
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0

    ${rc}  ${output}=  Run And Return Rc And Output  curl -u admin:Harbor12345 -s --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/v2.0/projects/library/repositories/${image}/artifacts/latest?with_signature=true"

    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  "signed":true

    Run Keyword If  ${with_remove} == ${true}  Remove Notary Signature  ${ip}  ${image}

Delete A Project Without Sign In Harbor
    [Arguments]  ${harbor_ip}=${ip}  ${username}=${HARBOR_ADMIN}  ${password}=${HARBOR_PASSWORD}
    ${d}=    Get Current Date    result_format=%m%s
    ${project_name}=  Set Variable  000${d}
    Create An New Project  ${project_name}
    Push Image  ${harbor_ip}  ${username}  ${password}  ${project_name}  hello-world
    Project Should Not Be Deleted  ${project_name}
    Go Into Project  ${project_name}
    Delete Repo  ${project_name}
    Navigate To Projects
    Project Should Be Deleted  ${project_name}

Manage Project Member Without Sign In Harbor
    [Arguments]  ${sign_in_user}  ${sign_in_pwd}  ${test_user1}=user005  ${test_user2}=user006  ${is_oidc_mode}=${false}
    ${d}=    Get current Date  result_format=%m%s
    Create An New Project  project${d}
    Push image  ip=${ip}  user=${sign_in_user}  pwd=${sign_in_pwd}  project=project${d}  image=hello-world
    Logout Harbor

    User Should Not Be A Member Of Project  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Manage Project Member  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Add  is_oidc_mode=${is_oidc_mode}
    User Should Be Guest  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Change User Role In Project  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Developer  is_oidc_mode=${is_oidc_mode}
    User Should Be Developer  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Change User Role In Project  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Admin  is_oidc_mode=${is_oidc_mode}
    User Should Be Admin  ${test_user1}  ${sign_in_pwd}  project${d}  ${test_user2}  is_oidc_mode=${is_oidc_mode}
    Change User Role In Project  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Master  is_oidc_mode=${is_oidc_mode}
    User Should Be Master  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Manage Project Member  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Remove  is_oidc_mode=${is_oidc_mode}
    User Should Not Be A Member Of Project  ${test_user1}  ${sign_in_pwd}  project${d}    is_oidc_mode=${is_oidc_mode}
    Push image  ip=${ip}  user=${sign_in_user}  pwd=${sign_in_pwd}  project=project${d}  image=hello-world
    User Should Be Guest  ${test_user2}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}

Helm CLI Push Without Sign In Harbor
    [Arguments]  ${sign_in_user}  ${sign_in_pwd}
    ${d}=   Get Current Date    result_format=%m%s
    Create An New Project  project${d}
    Helm Repo Add  ${HARBOR_URL}  ${sign_in_user}  ${sign_in_pwd}  project_name=project${d}
    Helm Repo Push  ${sign_in_user}  ${sign_in_pwd}  ${harbor_chart_filename}
    Go Into Project  project${d}  has_image=${false}
    Switch To Project Charts
    Go Into Chart Version  ${harbor_chart_name}
    Retry Wait Until Page Contains  ${harbor_chart_version}
    Capture Page Screenshot

Helm3 CLI Push Without Sign In Harbor
    [Arguments]  ${sign_in_user}  ${sign_in_pwd}
    ${d}=   Get Current Date    result_format=%m%s
    Create An New Project  project${d}
    Helm Repo Push  ${sign_in_user}  ${sign_in_pwd}  ${harbor_chart_filename}  helm_repo_name=${HARBOR_URL}/chartrepo/project${d}  helm_cmd=helm3
    Go Into Project  project${d}  has_image=${false}
    Switch To Project Charts
    Retry Double Keywords When Error  Go Into Chart Version  ${harbor_chart_name}  Retry Wait Until Page Contains  ${harbor_chart_version}
    Capture Page Screenshot

#Important Note: All CVE IDs in CVE Whitelist cases must unique!
Body Of Verfiy System Level CVE Whitelist
    [Arguments]  ${image_argument}  ${sha256_argument}  ${most_cve_list}  ${single_cve}
    [Tags]  run-once
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=    Set Variable    ${image_argument}
    # ${image}=    Set Variable    goharbor/harbor-portal
    ${sha256}=  Set Variable  ${sha256_argument}
    # ${sha256}=  Set Variable  2cb6a1c24dd6b88f11fd44ccc6560cb7be969f8ac5f752802c99cae6bcd592bb
    ${signin_user}=    Set Variable  user025
    ${signin_pwd}=    Set Variable  Test1@34
    Sign In Harbor    ${HARBOR_URL}    ${signin_user}    ${signin_pwd}
    Create An New Project    project${d}
    Push Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    sha256=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  2
    Cannot Pull image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image}
    Scan Repo  ${sha256}  Succeed
    Logout Harbor
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Switch To Configuration System Setting
    # Add Items To System CVE Whitelist    CVE-2019-19317\nCVE-2019-19646 \nCVE-2019-5188 \nCVE-2019-20387 \nCVE-2019-17498 \nCVE-2019-20372 \nCVE-2019-19244 \nCVE-2019-19603 \nCVE-2019-19880 \nCVE-2019-19923 \nCVE-2019-19925 \nCVE-2019-19926 \nCVE-2019-19959 \nCVE-2019-20218 \nCVE-2019-19232 \nCVE-2019-19234 \nCVE-2019-19645
    Add Items To System CVE Whitelist    ${most_cve_list}
    Cannot Pull image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    # Add Items To System CVE Whitelist    CVE-2019-18276
    Add Items To System CVE Whitelist    ${single_cve}
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Delete Top Item In System CVE Whitelist  count=6
    Cannot Pull image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Close Browser

Body Of Verfiy Project Level CVE Whitelist
    [Arguments]  ${image_argument}  ${sha256_argument}  ${most_cve_list}  ${single_cve}
    [Tags]  run-once
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=    Set Variable    ${image_argument}
    ${sha256}=  Set Variable  ${sha256_argument}
    ${signin_user}=    Set Variable  user025
    ${signin_pwd}=    Set Variable  Test1@34
    Sign In Harbor    ${HARBOR_URL}    ${signin_user}    ${signin_pwd}
    Create An New Project    project${d}
    Push Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    sha256=${sha256}
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  2
    Cannot Pull image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image}
    Scan Repo  ${sha256}  Succeed
    Go Into Project  project${d}
    Add Items to Project CVE Whitelist    ${most_cve_list}
    Cannot Pull image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Add Items to Project CVE Whitelist    ${single_cve}
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Delete Top Item In Project CVE Whitelist
    Cannot Pull image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Close Browser

Body Of Verfiy Project Level CVE Whitelist By Quick Way of Add System
    [Arguments]  ${image_argument}  ${sha256_argument}  ${cve_list}
    [Tags]  run-once
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${image}=    Set Variable    ${image_argument}
    ${sha256}=  Set Variable  ${sha256_argument}
    ${signin_user}=    Set Variable  user025
    ${signin_pwd}=    Set Variable  Test1@34
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Configure
    Switch To Configuration System Setting
    Add Items To System CVE Whitelist    ${cve_list}
    Logout Harbor
    Sign In Harbor    ${HARBOR_URL}    ${signin_user}    ${signin_pwd}
    Create An New Project    project${d}
    Push Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    sha256=${sha256}
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  2
    Go Into Project  project${d}
    Go Into Repo  project${d}/${image}
    Scan Repo  ${sha256}  Succeed
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Go Into Project  project${d}
    Set Project To Project Level CVE Whitelist
    Cannot Pull image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Add System CVE Whitelist to Project CVE Whitelist By Add System Button Click
    Pull Image    ${ip}    ${signin_user}    ${signin_pwd}    project${d}    ${image}    tag=${sha256}
    Close Browser