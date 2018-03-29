// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

*** Settings ***
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Default Tags  OVA

*** Test Cases ***
Test Case - Deploy OVA
    Deploy Harbor-OVA To Test Server  %{DHCP}  %{PROTOCOL}  False  %{USER}  %{PASSWORD}  ${ova_url}  %{HOST}  %{DATASTORE}  %{CLUSTER}  %{DATACENTER}

Test Case - Sign With Admin Modified Pwd
    Open Connection    %{HARBOR_IP}
    Login    root    ova-test-root-pwd
    SSHLibrary.Get File  /data/ca_download/harbor_ca.crt
    Close All Connections
    Generate Certificate Authority For Chrome  %{HARBOR_PASSWORD}	
    Init Chrome Driver
    Sign In Harbor  https://%{HARBOR_IP}  admin  %{HARBOR_ADMIN_PASSWORD}
    Close Browser
	
Test Case - Push Image
    Init Chrome Driver
    Start Docker Daemon Locally
    ${d}=    Get Current Date    result_format=%m%s
    Create An New User  url=https://%{HARBOR_IP}  username=tester${d}  email=tester${d}@vmware.com  realname=harbortest  newPassword=Test1@34  comment=harbortest
    Create An New Project  test${d}

    Push image  %{HARBOR_IP}  tester${d}  Test1@34  test${d}  hello-world:latest
    Go Into Project  test${d}
    Wait Until Page Contains  test${d}/hello-world
	
Test Case - OVA reboot
    Reboot VM  harbor-unified-ova-integration-test
    Wait for Harbor Ready  %{protocol}  %{HARBOR_IP}

Test Case - OVA reset
    Reset VM  harbor-unified-ova-integration-test
    Wait for Harbor Ready  %{protocol}  %{HARBOR_IP}	
