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
Documentation  This resource contains any keywords dealing with operations being performed on a Vsphere instance, mostly govc wrappers

*** Keywords ***
Power On VM OOB
    [Arguments]  ${vm}  ${vc_host}  ${vc_user}  ${vc_password}
    ${rc}  ${output}=  Run And Return Rc And Output  GOVC_URL=${vc_host} GOVC_USERNAME=${vc_user} GOVC_PASSWORD=${vc_password} GOVC_INSECURE=1 govc vm.power -on "${vm}"
    Should Be Equal As Integers  ${rc}  0
    Log To Console  Waiting for VM to power on ...
    Wait Until VM Powers On  "${vm}"  ${vc_host}  ${vc_user}  ${vc_password}

Power Off VM OOB
    [Arguments]  ${vm}  ${vc_host}  ${vc_user}  ${vc_password}
    ${rc}  ${output}=  Run And Return Rc And Output  GOVC_URL=${vc_host} GOVC_USERNAME=${vc_user} GOVC_PASSWORD=${vc_password} GOVC_INSECURE=1 govc vm.power -off "${vm}"
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    Log To Console  Waiting for VM to power off ...
    Wait Until VM Powers Off  "${vm}"  ${vc_host}  ${vc_user}  ${vc_password}

Destroy VM OOB
    [Arguments]  ${vm}
    ${rc}  ${output}=  Run And Return Rc And Output  govc vm.destroy "*-${vm}"
    Should Be Equal As Integers  ${rc}  0

Put Host Into Maintenance Mode
    ${rc}  ${output}=  Run And Return Rc And Output  govc host.maintenance.enter -host.ip=%{TEST_URL}
    Should Contain  ${output}  entering maintenance mode... OK

Remove Host From Maintenance Mode
    ${rc}  ${output}=  Run And Return Rc And Output  govc host.maintenance.exit -host.ip=%{TEST_URL}
    Should Contain  ${output}  exiting maintenance mode... OK

Reboot VM
    [Arguments]  ${vm}  ${vc_host}  ${vc_user}  ${vc_password}
    Log To Console  Rebooting ${vm} ...
    Power Off VM OOB  ${vm}  ${vc_host}  ${vc_user}  ${vc_password}
    Power On VM OOB  ${vm}  ${vc_host}  ${vc_user}  ${vc_password}
    Log To Console  ${vm} Powered On

Reset VM
    [Arguments]  ${vm}  ${vc_host}  ${vc_user}  ${vc_password}
    ${rc}  ${output}=  Run And Return Rc And Output  GOVC_URL=${vc_host} GOVC_USERNAME=${vc_user} GOVC_PASSWORD=${vc_password} GOVC_INSECURE=1 govc vm.power -reset "${vm}"
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    Log To Console  Waiting for VM to reset ...
    Wait Until VM Powers On  "${vm}"  ${vc_host}  ${vc_user}  ${vc_password}

Wait Until VM Powers On
    [Arguments]  ${vm}  ${vc_host}  ${vc_user}  ${vc_password}
    :FOR  ${idx}  IN RANGE  0  30
    \   ${ret}=  Run  GOVC_URL=${vc_host} GOVC_USERNAME=${vc_user} GOVC_PASSWORD=${vc_password} GOVC_INSECURE=1 govc vm.info ${vm}
    \   Set Test Variable  ${out}  ${ret}
    \   ${status}=  Run Keyword And Return Status  Should Contain  ${out}  poweredOn
    \   Return From Keyword If  ${status}
    \   Sleep  1
    Fail  VM did not power on within 30 seconds

Wait Until VM Powers Off
    [Arguments]  ${vm}  ${vc_host}  ${vc_user}  ${vc_password}
    :FOR  ${idx}  IN RANGE  0  30
    \   ${ret}=  Run  GOVC_URL=${vc_host} GOVC_USERNAME=${vc_user} GOVC_PASSWORD=${vc_password} GOVC_INSECURE=1 govc vm.info ${vm}
    \   Set Test Variable  ${out}  ${ret}
    \   ${status}=  Run Keyword And Return Status  Should Contain  ${out}  poweredOff
    \   Return From Keyword If  ${status}
    \   Sleep  1
    Fail  VM did not power off within 30 seconds

Wait Until VM Is Destroyed
    [Arguments]  ${vm}
    :FOR  ${idx}  IN RANGE  0  30
    \   ${ret}=  Run  govc ls vm/${vm}
    \   Set Test Variable  ${out}  ${ret}
    \   ${status}=  Run Keyword And Return Status  Should Be Empty  ${out}
    \   Return From Keyword If  ${status}
    \   Sleep  1
    Fail  VM was not destroyed within 30 seconds

Get VM IP
    [Arguments]  ${vm}
    ${rc}  ${out}=  Run And Return Rc And Output  govc vm.ip ${vm}
    Should Be Equal As Integers  ${rc}  0
    [Return]  ${out}

Get VM Host Name
    [Arguments]  ${vm}
    ${out}=  Run  govc vm.info ${vm}
    ${out}=  Split To Lines  ${out}
    ${host}=  Fetch From Right  @{out}[-1]  ${SPACE}
    [Return]  ${host}

Get VM Info
    [Arguments]  ${vm}
    ${rc}  ${out}=  Run And Return Rc And Output  govc vm.info -r ${vm}
    Should Be Equal As Integers  ${rc}  0
    [Return]  ${out}

vMotion A VM
    [Arguments]  ${vm}
    ${host}=  Get VM Host Name  ${vm}
    ${status}=  Run Keyword And Return Status  Should Contain  ${host}  ${esx1-ip}
    Run Keyword If  ${status}  Run  govc vm.migrate -host cls/${esx2-ip} -pool cls/Resources ${vm}
    Run Keyword Unless  ${status}  Run  govc vm.migrate -host cls/${esx1-ip} -pool cls/Resources ${vm}

Create Test Server Snapshot
    [Arguments]  ${vm}  ${snapshot}
    Set Environment Variable  GOVC_URL  %{BUILD_SERVER}
    ${rc}  ${out}=  Run And Return Rc And Output  govc snapshot.create -vm ${vm} ${snapshot}
    Should Be Equal As Integers  ${rc}  0
    Should Be Empty  ${out}
    Set Environment Variable  GOVC_URL  %{TEST_USERNAME}:%{TEST_PASSWORD}@%{TEST_URL}

Revert Test Server Snapshot
    [Arguments]  ${vm}  ${snapshot}
    Set Environment Variable  GOVC_URL  %{BUILD_SERVER}
    ${rc}  ${out}=  Run And Return Rc And Output  govc snapshot.revert -vm ${vm} ${snapshot}
    Should Be Equal As Integers  ${rc}  0
    Should Be Empty  ${out}
    Set Environment Variable  GOVC_URL  %{TEST_USERNAME}:%{TEST_PASSWORD}@%{TEST_URL}

Delete Test Server Snapshot
    [Arguments]  ${vm}  ${snapshot}
    Set Environment Variable  GOVC_URL  %{BUILD_SERVER}
    ${rc}  ${out}=  Run And Return Rc And Output  govc snapshot.remove -vm ${vm} ${snapshot}
    Should Be Equal As Integers  ${rc}  0
    Should Be Empty  ${out}
    Set Environment Variable  GOVC_URL  %{TEST_USERNAME}:%{TEST_PASSWORD}@%{TEST_URL}

Setup Snapshot
    ${hostname}=  Get Test Server Hostname
    Set Environment Variable  TEST_HOSTNAME  ${hostname}
    Set Environment Variable  SNAPSHOT  vic-ci-test-%{DRONE_BUILD_NUMBER}
    Create Test Server Snapshot  %{TEST_HOSTNAME}  %{SNAPSHOT}

Get Datacenter Name
    ${out}=  Run  govc datacenter.info
    ${out}=  Split To Lines  ${out}
    ${name}=  Fetch From Right  @{out}[0]  ${SPACE}
    [Return]  ${name}

Get Test Server Hostname
    [Tags]  secret
    ${hostname}=  Run  sshpass -p $TEST_PASSWORD ssh $TEST_USERNAME@$TEST_URL hostname
    [Return]  ${hostname}

Check Delete Success
    [Arguments]  ${name}
    ${out}=  Run  govc ls vm
    Log  ${out}
    Should Not Contain  ${out}  ${name}
    ${out}=  Run  govc datastore.ls
    Log  ${out}
    Should Not Contain  ${out}  ${name}
    ${out}=  Run  govc ls host/*/Resources/*
    Log  ${out}
    Should Not Contain  ${out}  ${name}

Gather Logs From ESX Server
    Environment Variable Should Be Set  TEST_URL
    ${out}=  Run  govc logs.download

Change Log Level On Server
    [Arguments]  ${level}
    ${out}=  Run  govc host.option.set Config.HostAgent.log.level ${level}
    Should Be Empty  ${out}

Add Vsphere License
    [Tags]  secret
    [Arguments]  ${license}
    ${out}=  Run  govc license.add ${license}
    Should Contain  ${out}  Key:

Assign Vsphere License
    [Tags]  secret
    [Arguments]  ${license}  ${host}
    ${out}=  Run  govc license.assign -host ${host} ${license}
    Should Contain  ${out}  Key:

Add Host To VCenter
    [Arguments]  ${host}  ${user}  ${dc}  ${pw}
    :FOR  ${idx}  IN RANGE  1  4
    \   ${out}=  Run  govc cluster.add -hostname=${host} -username=${user} -dc=${dc} -password=${pw} -noverify=true
    \   ${status}=  Run Keyword And Return Status  Should Contain  ${out}  OK
    \   Return From Keyword If  ${status}
    Fail  Failed to add the host to the VC in 3 attempts

Get Host Firewall Enabled
    ${output}=  Run  govc host.esxcli network firewall get
    Should Contain  ${output}  Enabled
    @{output}=  Split To Lines  ${output}
    :FOR  ${line}  IN  @{output}
    \   Run Keyword If  "Enabled" in '''${line}'''  Set Test Variable  ${out}  ${line}
    ${enabled}=  Fetch From Right  ${out}  :
    ${enabled}=  Strip String  ${enabled}
    Return From Keyword If  '${enabled}' == 'false'  ${false}
    Return From Keyword If  '${enabled}' == 'true'  ${true}

Enable Host Firewall
    Run  govc host.esxcli network firewall set --enabled true

Disable Host Firewall
    Run  govc host.esxcli network firewall set --enabled false