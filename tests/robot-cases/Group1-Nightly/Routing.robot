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
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Main Menu Routing
    [Tags]  main_menu_routing
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    &{routing}=	 Create Dictionary  harbor/projects=//projects//div//h2[contains(.,'Projects')]
    ...  harbor/logs=//hbr-log//div//h2[contains(.,'Logs')]
    ...  harbor/users=//harbor-user//div//h2[contains(.,'Users')]
    ...  harbor/robot-accounts=//system-robot-accounts//h2[contains(.,'Robot Accounts')]
    ...  harbor/registries=//hbr-endpoint//h2[contains(.,'Registries')]
    ...  harbor/replications=//total-replication//h2[contains(.,'Replications')]
    ...  harbor/distribution/instances=//dist-instances//div//h2[contains(.,'Instances')]
    ...  harbor/labels=//app-labels//h2[contains(.,'Labels')]
    ...  harbor/project-quotas=//app-project-quotas//h2[contains(.,'Project Quotas')]
    ...  harbor/interrogation-services/scanners=//config-scanner//div//h4[contains(.,'Image Scanners')]
    ...  harbor/interrogation-services/vulnerability=//vulnerability-config//div//button[contains(.,'SCAN NOW')]
    ...  harbor/gc=//app-gc-page//h2[contains(.,'Garbage Collection')]
    ...  harbor/configs/auth=//config//config-auth//label[contains(.,'Auth Mode')]
    ...  /harbor/configs/email=//config//config-email//label[contains(.,'Email Server Port')]
    ...  harbor/configs/setting=//config//system-settings//label[contains(.,'Project Creation')]
    FOR  ${key}  IN  @{routing.keys()}
        Retry Double Keywords When Error  Go To  ${HARBOR_URL}/${key}  Retry Wait Element  ${routing['${key}']}
    END

Test Case - Project Tab Routing
    [Tags]  project_tab_routing
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    &{routing}=	 Create Dictionary  harbor/projects/1/summary=//project-detail//summary
    ...  harbor/projects/1/repositories=//project-detail//hbr-repository-gridview
    ...  harbor/projects/1/helm-charts=//project-detail//project-list-charts
    ...  harbor/projects/1/members=//project-detail//ng-component//button//span[contains(.,'User')]
    ...  harbor/projects/1/labels=//project-detail//app-project-config//hbr-label
    ...  harbor/projects/1/scanner=//project-detail//scanner
    ...  harbor/projects/1/p2p-provider/policies=//project-detail//ng-component//button//span[contains(.,'NEW POLICY')]
    ...  harbor/projects/1/tag-strategy/tag-retention=//project-detail//app-tag-feature-integration//tag-retention
    ...  harbor/projects/1/tag-strategy/immutable-tag=//project-detail//app-tag-feature-integration//app-immutable-tag
    ...  harbor/projects/1/robot-account=//project-detail//app-robot-account
    ...  harbor/projects/1/webhook=//project-detail//ng-component//button//span[contains(.,'New Webhook')]
    ...  harbor/projects/1/logs=//project-detail//audit-log
    ...  harbor/projects/1/configs=//project-detail//app-project-config//hbr-project-policy-config
    FOR  ${key}  IN  @{routing.keys()}
        Retry Double Keywords When Error  Go To  ${HARBOR_URL}/${key}  Retry Wait Element  ${routing['${key}']}
    END