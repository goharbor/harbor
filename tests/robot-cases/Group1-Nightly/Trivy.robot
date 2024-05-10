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
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${ip}
${SSH_USER}  root
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Get Harbor Version
#Just get harbor version and log it
    Get Harbor Version

Test Case - Trivy Is Default Scanner And It Is Immutable
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Scanners Page
    Should Display The Default Trivy Scanner
    Trivy Is Immutable Scanner

Test Case - Disable Scan Schedule
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Vulnerability Page
    Disable Scan Schedule
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Vulnerability Page
    Retry Wait Until Page Contains  None
    Close Browser

Test Case - Security Hub
    [Tags]  security_hub
    ${d}=  Get Current Date  result_format=%m%s
    ${images}=  Create List  goharbor/harbor-log-base  goharbor/harbor-prepare-base  goharbor/harbor-redis-base  goharbor/harbor-nginx-base  goharbor/harbor-registry-base
    ${tag}=  Set Variable  v2.2.0
    ${digest}=  Set Variable  sha256:7bf979f25c6a6986eab83e100a7b78bd5195c9bcac03e823e64492bb17fa4dad
    ${cve_id}=  Set Variable  CVE-2021-22926
    ${package}=  Set Variable  curl
    ${cvss_score_v3_from}=  Set Variable  6.5
    ${cvss_score_v3_to}=  Set Variable  7.5
    ${severity}=  Set Variable  High
    ${index_repo}=  Set Variable  index${d}
    ${cve_description}=  Set Variable  Description: libcurl-using applications can ask for a specific client certificate to be used in a transfer.
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    FOR  ${image}  IN  @{images}
        Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}  ${tag}  ${tag}
        Go Into Repo  project${d}  ${image}
        Scan Repo  ${tag}  Succeed
    END
    Docker Push Index  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${ip}/project${d}/${index_repo}:${tag}  ${ip}/project${d}/${images}[0]:${tag}  ${ip}/project${d}/${images}[1]:${tag}
    Back Project Home  project${d}
    Refresh Repositories
    Delete Repo  project${d}  ${images}[0]
    Delete Repo  project${d}  ${images}[1]
    Go Into Repo  project${d}  ${index_repo}
    Scan Repo  ${tag}  Succeed
    Switch To Security Hub
    ${summary}=  Get Vulnerability System Summary From API
    # Check The Total Vulnerabilities
    Check The Total Vulnerabilities  ${summary}
    # Check the Top 5 Most Dangerous Artifacts
    ${dangerous_artifacts}=  Set Variable  ${summary["dangerous_artifacts"]}
    Check The Top 5 Most Dangerous Artifacts  ${dangerous_artifacts}
    # Check the Top 5 Most Dangerous CVEs
    ${dangerous_cves}=  Set Variable  ${summary["dangerous_cves"]}
    Check The Top 5 Most Dangerous CVEs  ${dangerous_cves}
    # Check the vulnerabilities search
    Retry Wait Element Not Visible  ${add_search_criteria_icon}
    Retry Wait Element Not Visible  ${remove_search_criteria_icon}
    Retry Wait Element Count  ${vulnerabilities_datagrid_row}  10
    Check The Quick Search
    Check The Search By One Condition  project${d}  project${d}/${images}[2]  ${digest}  ${cve_id}  ${package}  ${tag}  ${cvss_score_v3_from}  ${cvss_score_v3_to}  ${summary}
    Check The Search By All Condition  project${d}  project${d}/${images}[2]  ${digest}  ${cve_id}  ${package}  ${tag}  ${cvss_score_v3_from}  ${cvss_score_v3_to}  ${severity}
    # Check the vulnerabilities jump
    Check The Vulnerabilities Jump  project${d}  ${images}[2]  ${cve_id}  ${cve_description}
    # Check that there is no such artifact in the security hub after deleting the artifact
    Go Into Project  project${d}
    Delete Repo  project${d}  ${index_repo}
    Switch To Security Hub
    Retry Wait Until Page Not Contains Element  //div[@class='card'][2]//a[@title='project${d}/${index_repo}']
    Select From List By Value  ${vulnerabilities_filter_select}  repository_name
    Retry Text Input  ${vulnerabilities_filter_input}  project${d}/${index_repo}
    Retry Button Click  ${security_hub_search_btn}
    Retry Wait Element Count  //div[@class='datagrid']//clr-dg-cell[2][@title='project${d}/${index_repo}']  0
    Close Browser

Test Case - Manual Scan All
    Body Of Manual Scan All  Critical  High

Test Case - Scan A Tag In The Repo
    Body Of Scan A Tag In The Repo  vmware/photon  1.0

Test Case - Scan As An Unprivileged User
    Init Chrome Driver
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  hello-world
    Sign In Harbor  ${HARBOR_URL}  user024  Test1@34
    Go Into Repo  library  hello-world
    Select Object  latest
    Scan Is Disabled
    Close Browser

# Chose a empty Vul repo
Test Case - Scan Image With Empty Vul
    Body Of Scan Image With Empty Vul  photon  4.0_scan

Test Case - Scan Image On Push
    [Tags]  run-once
    Body Of Scan Image On Push  Critical  High

Test Case - View Scan Results
    [Tags]  run-once
    Body Of View Scan Results  Critical

Test Case - Project Level Image Serverity Policy
    [Tags]  run-once
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    ${d}=  get current date  result_format=%m%s
    ${sha256}=  Set Variable  0e67625224c1da47cb3270e7a861a83e332f708d3d89dde0cbed432c94824d9a
    ${image}=  Set Variable  redis
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}  sha256=${sha256}
    Go Into Repo  project${d}  ${image}
    Scan Repo  ${sha256}  Succeed
    Navigate To Projects
    Go Into Project  project${d}
    Set Vulnerabilty Serverity  3
    Cannot Pull Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}  tag=${sha256}  err_msg=To continue with pull, please contact your project administrator to exempt matched vulnerabilities through configuring the CVE allowlist
    Close Browser

#Important Note: All CVE IDs in CVE Allowlist cases must unique!
Test Case - Verfiy System Level CVE Allowlist
    [Tags]  sys_cve
    Body Of Verfiy System Level CVE Allowlist  goharbor/harbor-portal  55d776fc7f431cdd008c3d8fc3e090c81c1368ed9ed85335f4664df71f864f0d  CVE-2021-36222\nCVE-2021-43527 \nCVE-2021-4044 \nCVE-2021-36084 \nCVE-2021-36085 \nCVE-2021-36086 \nCVE-2021-37750 \nCVE-2021-40528  CVE-2021-43519

Test Case - Verfiy Project Level CVE Allowlist
    [Tags]  proj_cve
    Body Of Verfiy Project Level CVE Allowlist  goharbor/harbor-portal  55d776fc7f431cdd008c3d8fc3e090c81c1368ed9ed85335f4664df71f864f0d  CVE-2021-36222\nCVE-2021-43527 \nCVE-2021-4044 \nCVE-2021-36084 \nCVE-2021-36085 \nCVE-2021-36086 \nCVE-2021-37750 \nCVE-2021-40528  CVE-2021-43519

Test Case - Verfiy Project Level CVE Allowlist By Quick Way of Add System
    [Tags]  proj_cve_quick_add_sys
    Body Of Verfiy Project Level CVE Allowlist By Quick Way of Add System  goharbor/harbor-portal  55d776fc7f431cdd008c3d8fc3e090c81c1368ed9ed85335f4664df71f864f0d  CVE-2021-36222\nCVE-2021-43527 \nCVE-2021-4044 \nCVE-2021-36084 \nCVE-2021-36085 \nCVE-2021-36086 \nCVE-2021-37750 \nCVE-2021-40528 \nCVE-2021-43519

Test Case - Stop Scan And Stop Scan All
    [Tags]  stop_scan_job
    Body Of Stop Scan And Stop Scan All

Test Case - External Scanner CRUD
    [Tags]  external_scanner_crud  need_scanner_endpoint
    ${SCANNER_ENDPOINT_VALUE}=  Get Variable Value  ${SCANNER_ENDPOINT}  ${EMPTY}
    Skip If  '${SCANNER_ENDPOINT_VALUE}' == '${EMPTY}'
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To Scanners Page
    # Add a new scanner
    Add A New Scanner  scanner${d}  ${SCANNER_ENDPOINT}  Basic  For testing  ${true}  ${true}  scanner_name  scanner_password
    # Update this scanner
    Update Scanner  scanner${d}  scanner${d}-edit1  ${SCANNER_ENDPOINT}1  Bearer  For testing-edit1  ${true}  ${true}  token=scanner_token
    Update Scanner  scanner${d}-edit1  scanner${d}-edit2  ${SCANNER_ENDPOINT}2  APIKey  For testing-edit2  ${true}  ${true}  api_key=scanner_api_key
    Update Scanner  scanner${d}-edit2  scanner${d}  ${SCANNER_ENDPOINT}  None  For testing
    # Filter this scanner
    Filter Scanner By Name  scanner${d}
    Filter Scanner By Endpoint  ${SCANNER_ENDPOINT}
    Retry Wait Element Count  //clr-dg-row  1
    Retry Double Keywords When Error  Retry Element Click  xpath=${scanner_list_refresh_btn}  Retry Wait Until Page Contains Element  //clr-dg-row[.//span[text()='scanner${d}'] and .//clr-dg-cell[text()='${SCANNER_ENDPOINT}'] and .//span[text()='Healthy'] and .//clr-dg-cell[text()='None']]
    # Delete this scanner
    Delete Scanner  scanner${d}
    Close Browser

Test Case - Set External Scanner As Default And Scan
    [Tags]  external_scanner_scan  need_scanner_endpoint
    ${SCANNER_ENDPOINT_VALUE}=  Get Variable Value  ${SCANNER_ENDPOINT}  ${EMPTY}
    Skip If  '${SCANNER_ENDPOINT_VALUE}' == '${EMPTY}'
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s
    ${image}=  Set Variable  hello-world
    ${tag}=  Set Variable  latest
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Switch To Project Scanner
    Retry Wait Element Visible  //span[@id='scanner-name' and text()='Trivy']
    Switch To Scanners Page
    Retry Wait Element Visible  //clr-dg-row[.//span[text()='Trivy'] and .//span[text()='Default']]
    # Add a new scanner
    Add A New Scanner  scanner${d}  ${SCANNER_ENDPOINT}  None  For testing
    # Set this scanner to default
    Set Scanner As Default  scanner${d}
    Go Into Project  project${d}  ${false}
    Switch To Project Scanner
    Retry Wait Element Visible  //span[@id='scanner-name' and text()='scanner${d}']
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}
    Go Into Repo  project${d}  ${image}
    Scan Repo  ${tag}  Succeed
    Retry Wait Element Visible  //hbr-result-tip-histogram//span[1][text()='10']
    Switch To Scanners Page
    Set Scanner As Default  Trivy
    Go Into Project  project${d}
    Switch To Project Scanner
    Retry Wait Element Visible  //span[@id='scanner-name' and text()='Trivy']
    Go into Repo  project${d}  ${image}
    Scan Repo  ${tag}  Succeed
    Retry Wait Element Visible  //hbr-result-tip-histogram//div[text()=' No vulnerability ']
    Back Project Home  project${d}
    Switch To Project Scanner
    Select Project Scanner  scanner${d}  2
    Go Into Repo  project${d}  ${image}
    Retry Wait Element Visible  //hbr-result-tip-histogram//span[1][text()='10']
    Scan Repo  ${tag}  Succeed
    Retry Wait Element Visible  //hbr-result-tip-histogram//span[1][text()='10']
    Back Project Home  project${d}
    Switch To Project Scanner
    Select Project Scanner  Trivy  2
    Go Into Repo  project${d}  ${image}
    Retry Wait Element Visible  //hbr-result-tip-histogram//div[text()=' No vulnerability ']
    Scan Repo  ${tag}  Succeed
    Retry Wait Element Visible  //hbr-result-tip-histogram//div[text()=' No vulnerability ']
    Switch To Scanners Page
    Delete Scanner  scanner${d}
    Go Into Project  project${d}
    Switch To Project Scanner
    Retry Element Click  //*[@id='edit-scanner']
    Retry Wait Element Count  //clr-dg-row  1
    Close Browser

Test Case - Enable And Deactivate Scanner
    [Tags]  enable_deactivate_scanner  need_scanner_endpoint
    ${SCANNER_ENDPOINT_VALUE}=  Get Variable Value  ${SCANNER_ENDPOINT}  ${EMPTY}
    Skip If  '${SCANNER_ENDPOINT_VALUE}' == '${EMPTY}'
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s
    ${image}=  Set Variable  hello-world
    ${tag}=  Set Variable  latest
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project And Go Into Project  project${d}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}
    Switch To Scanners Page
    Add A New Scanner  scanner${d}  ${SCANNER_ENDPOINT}  None  For testing
    Go Into Project  project${d}
    Switch To Project Scanner
    Select Project Scanner  scanner${d}
    Go into Repo  project${d}  ${image}
    Scan Repo  ${tag}  Succeed
    # Deactivate this scanner
    Switch To Scanners Page
    Enable Or Deactivate Scanner  scanner${d}  DEACTIVATE
    Go Into Project  project${d}
    Switch To Project Scanner
    Retry Wait Element Visible  //scanner//span[text()='Deactivated']
    Go Into Repo  project${d}  ${image}
    Retry Element Click  //clr-dg-row[contains(.,'${tag}')]//label[contains(@class,'clr-control-label')]
    Retry Wait Element Should Be Disabled  //button[@id='scan-btn']
    # Enable this scanner
    Switch To Scanners Page
    Enable Or Deactivate Scanner  scanner${d}  ENABLE
    Go Into Project  project${d}
    Switch To Project Scanner
    Retry Wait Element Not Visible  //scanner//span[text()='Deactivated']
    Go Into Repo  project${d}  ${image}
    Scan Repo  ${tag}  Succeed
    Switch To Scanners Page
    Delete Scanner  scanner${d}
    Go Into Project  project${d}
    Switch To Project Scanner
    Retry Wait Element Visible  //span[@id='scanner-name' and text()='Trivy']
    Close Browser
