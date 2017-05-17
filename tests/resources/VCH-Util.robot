# Copyright 2016-2017 VMware, Inc. All Rights Reserved.
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
Documentation  This resource contains all keywords related to creating, deleting, maintaining VCHs

*** Keywords ***
Set Test Environment Variables
    # Finish setting up environment variables
    ${status}  ${message}=  Run Keyword And Ignore Error  Environment Variable Should Be Set  DRONE_BUILD_NUMBER
    Run Keyword If  '${status}' == 'FAIL'  Set Environment Variable  DRONE_BUILD_NUMBER  0
    ${status}  ${message}=  Run Keyword And Ignore Error  Environment Variable Should Be Set  BRIDGE_NETWORK
    Run Keyword If  '${status}' == 'FAIL'  Set Environment Variable  BRIDGE_NETWORK  network
    ${status}  ${message}=  Run Keyword And Ignore Error  Environment Variable Should Be Set  PUBLIC_NETWORK
    Run Keyword If  '${status}' == 'FAIL'  Set Environment Variable  PUBLIC_NETWORK  'VM Network'
    ${status}  ${message}=  Run Keyword And Ignore Error  Environment Variable Should Be Set  TEST_DATACENTER
    Run Keyword If  '${status}' == 'FAIL'  Set Environment Variable  TEST_DATACENTER  ${SPACE}

    @{URLs}=  Split String  %{TEST_URL_ARRAY}
    ${len}=  Get Length  ${URLs}
    ${IDX}=  Evaluate  %{DRONE_BUILD_NUMBER} \% ${len}

    Set Environment Variable  TEST_URL  @{URLs}[${IDX}]
    Set Environment Variable  GOVC_URL  %{TEST_USERNAME}:%{TEST_PASSWORD}@%{TEST_URL}
    # TODO: need an integration/vic-test image update to include the about.cert command
    #${rc}  ${thumbprint}=  Run And Return Rc And Output  govc about.cert -k | jq -r .ThumbprintSHA1
    ${rc}  ${thumbprint}=  Run And Return Rc And Output  openssl s_client -connect $(govc env -x GOVC_URL_HOST):443 </dev/null 2>/dev/null | openssl x509 -fingerprint -noout | cut -d= -f2
    Should Be Equal As Integers  ${rc}  0
    Set Environment Variable  TEST_THUMBPRINT  ${thumbprint}
    Log To Console  \nTEST_URL=%{TEST_URL}

    ${host}=  Run  govc ls host
    ${status}  ${message}=  Run Keyword And Ignore Error  Environment Variable Should Be Set  TEST_RESOURCE
    Run Keyword If  '${status}' == 'FAIL'  Set Environment Variable  TEST_RESOURCE  ${host}/Resources
    Set Environment Variable  GOVC_RESOURCE_POOL  %{TEST_RESOURCE}
    ${noQuotes}=  Strip String  %{TEST_DATASTORE}  characters="
    Set Environment Variable  GOVC_DATASTORE  ${noQuotes}

    ${about}=  Run  govc about
    ${status}=  Run Keyword And Return Status  Should Contain  ${about}  VMware ESXi
    Run Keyword If  ${status}  Set Environment Variable  HOST_TYPE  ESXi
    Run Keyword Unless  ${status}  Set Environment Variable  HOST_TYPE  VC

    ${about}=  Run  govc datastore.info %{TEST_DATASTORE} | grep 'Type'
    ${status}=  Run Keyword And Return Status  Should Contain  ${about}  vsan
    Run Keyword If  ${status}  Set Environment Variable  DATASTORE_TYPE  VSAN
    Run Keyword Unless  ${status}  Set Environment Variable  DATASTORE_TYPE  Non_VSAN

    # set the TLS config options suitable for vic-machine in this env
    ${domain}=  Get Environment Variable  DOMAIN  ''
    Run Keyword If  $domain == ''  Set Suite Variable  ${vicmachinetls}  --no-tlsverify
    Run Keyword If  $domain != ''  Set Suite Variable  ${vicmachinetls}  --tls-cname=*.${domain}

    Set Test VCH Name
    # Set a unique bridge network for each VCH that has a random VLAN ID
    ${vlan}=  Run Keyword If  '%{HOST_TYPE}' == 'ESXi'  Evaluate  str(random.randint(1, 4093))  modules=random
    ${out}=  Run Keyword If  '%{HOST_TYPE}' == 'ESXi'  Run  govc host.portgroup.add -vlan=${vlan} -vswitch vSwitchLAN %{VCH-NAME}-bridge
    Run Keyword If  '%{HOST_TYPE}' == 'ESXi'  Set Environment Variable  BRIDGE_NETWORK  %{VCH-NAME}-bridge

Set Test VCH Name
    ${name}=  Evaluate  'VCH-%{DRONE_BUILD_NUMBER}-' + str(random.randint(1000,9999))  modules=random
    Set Environment Variable  VCH-NAME  ${name}

Set List Of Env Variables
    [Arguments]  ${vars}
    @{vars}=  Split String  ${vars}
    :FOR  ${var}  IN  @{vars}
    \   ${varname}  ${varval}=  Split String  ${var}  =
    \   Set Environment Variable  ${varname}  ${varval}

Parse Environment Variables
    [Arguments]  ${line}
    #  If using the old logging format
    ${status}=  Run Keyword And Return Status  Should Contain  ${line}  mINFO
    ${logdeco}  ${vars}=  Run Keyword If  ${status}  Split String  ${line}  ${SPACE}  1
    Run Keyword If  ${status}  Set List Of Env Variables  ${vars}
    Return From Keyword If  ${status}

    # Split the log log into pieces, discarding the initial log decoration, and assign to env vars
    ${logmon}  ${logday}  ${logyear}  ${logtime}  ${loglevel}  ${vars}=  Split String  ${line}  max_split=5
    Set List Of Env Variables  ${vars}

Get Docker Params
    # Get VCH docker params e.g. "-H 192.168.218.181:2376 --tls"
    [Arguments]  ${output}  ${certs}
    @{output}=  Split To Lines  ${output}
    :FOR  ${item}  IN  @{output}
    \   ${status}  ${message}=  Run Keyword And Ignore Error  Should Contain  ${item}  DOCKER_HOST=
    \   Run Keyword If  '${status}' == 'PASS'  Set Suite Variable  ${line}  ${item}

    # Ensure we start from a clean slate with docker env vars
    Remove Environment Variable  DOCKER_HOST  DOCKER_TLS_VERIFY  DOCKER_CERT_PATH  CURL_CA_BUNDLE  COMPOSE_PARAMS  COMPOSE_TLS_VERSION

    Parse Environment Variables  ${line}

    ${dockerHost}=  Get Environment Variable  DOCKER_HOST

    @{hostParts}=  Split String  ${dockerHost}  :
    ${ip}=  Strip String  @{hostParts}[0]
    ${port}=  Strip String  @{hostParts}[1]
    Set Environment Variable  VCH-IP  ${ip}
    Set Environment Variable  VCH-PORT  ${port}

    :FOR  ${index}  ${item}  IN ENUMERATE  @{output}
    \   ${status}  ${message}=  Run Keyword And Ignore Error  Should Contain  ${item}  http
    \   Run Keyword If  '${status}' == 'PASS'  Set Suite Variable  ${line}  ${item}
    \   ${status}  ${message}=  Run Keyword And Ignore Error  Should Contain  ${item}  Published ports can be reached at
    \   ${idx} =  Evaluate  ${index} + 1
    \   Run Keyword If  '${status}' == 'PASS'  Set Suite Variable  ${ext-ip}  @{output}[${idx}]

    ${rest}  ${ext-ip} =  Split String From Right  ${ext-ip}  ${SPACE}  1
    ${ext-ip} =  Strip String  ${ext-ip}
    Set Environment Variable  EXT-IP  ${ext-ip}

    ${rest}  ${vic-admin}=  Split String From Right  ${line}  ${SPACE}  1
    Set Environment Variable  VIC-ADMIN  ${vic-admin}

    Run Keyword If  ${port} == 2376  Set Environment Variable  VCH-PARAMS  -H ${dockerHost} --tls
    Run Keyword If  ${port} == 2375  Set Environment Variable  VCH-PARAMS  -H ${dockerHost}

    ### Add environment variables for Compose and TLS

    # Check if tls is enable from vic-machine's output and not trust ${certs} which some tests bypasses
    ${tls_enabled}=  Get Environment Variable  DOCKER_TLS_VERIFY  ${false}

    ### Compose case for no-tlsverify

    # Set environment variables if certs not used to create the VCH.  This is NOT the recommended
    # approach to running compose.  There will be security warnings in the logs and some compose
    # operations may not work properly because certs == false currently means we install with
    # --no-tlsverify. Add CURL_CA_BUNDLE for a workaround in compose tests.  If we change
    # certs == false to install with --no-tls, then we need to change this again.
    Run Keyword If  ${tls_enabled} == ${false}  Set Environment Variable  CURL_CA_BUNDLE  ${EMPTY}

    # Get around quirk in compose if no-tlsverify, then CURL_CA_BUNDLE must exist and compose called with --tls
    Run Keyword If  ${tls_enabled} == ${false}  Set Environment Variable  COMPOSE-PARAMS  -H ${dockerHost} --tls

    ### Compose case for tlsverify (assumes DOCKER_TLS_VERIFY also set)

    Run Keyword If  ${tls_enabled} == ${true}  Set Environment Variable  COMPOSE_TLS_VERSION  TLSv1_2
    Run Keyword If  ${tls_enabled} == ${true}  Set Environment Variable  COMPOSE-PARAMS  -H ${dockerHost}

Install VIC Appliance To Test Server
    [Arguments]  ${vic-machine}=bin/vic-machine-linux  ${appliance-iso}=bin/appliance.iso  ${bootstrap-iso}=bin/bootstrap.iso  ${certs}=${true}  ${vol}=default  ${cleanup}=${true}
    Set Test Environment Variables
    # disable firewall
    Run Keyword If  '%{HOST_TYPE}' == 'ESXi'  Run  govc host.esxcli network firewall set -e false
    # Attempt to cleanup old/canceled tests
    Run Keyword If  ${cleanup}  Run Keyword And Ignore Error  Cleanup Dangling VMs On Test Server
    Run Keyword If  ${cleanup}  Run Keyword And Ignore Error  Cleanup Datastore On Test Server
    Run Keyword If  ${cleanup}  Run Keyword And Ignore Error  Cleanup Dangling Networks On Test Server
    Run Keyword If  ${cleanup}  Run Keyword And Ignore Error  Cleanup Dangling vSwitches On Test Server
    Run Keyword If  ${cleanup}  Run Keyword And Ignore Error  Cleanup Dangling Containers On Test Server

    # Install the VCH now
    Log To Console  \nInstalling VCH to test server...
    ${output}=  Run VIC Machine Command  ${vic-machine}  ${appliance-iso}  ${bootstrap-iso}  ${certs}  ${vol}
    Log  ${output}
    Should Contain  ${output}  Installer completed successfully
    Get Docker Params  ${output}  ${certs}
    Log To Console  Installer completed successfully: %{VCH-NAME}...

Run VIC Machine Command
    [Tags]  secret
    [Arguments]  ${vic-machine}  ${appliance-iso}  ${bootstrap-iso}  ${certs}  ${vol}
    ${output}=  Run Keyword If  ${certs}  Run  ${vic-machine} create --debug 1 --name=%{VCH-NAME} --target=%{TEST_URL}%{TEST_DATACENTER} --thumbprint=%{TEST_THUMBPRINT} --user=%{TEST_USERNAME} --image-store=%{TEST_DATASTORE} --appliance-iso=${appliance-iso} --bootstrap-iso=${bootstrap-iso} --password=%{TEST_PASSWORD} --force=true --bridge-network=%{BRIDGE_NETWORK} --public-network=%{PUBLIC_NETWORK} --compute-resource=%{TEST_RESOURCE} --timeout %{TEST_TIMEOUT} --volume-store=%{TEST_DATASTORE}/test:${vol} ${vicmachinetls}
    Run Keyword If  ${certs}  Should Contain  ${output}  Installer completed successfully
    Return From Keyword If  ${certs}  ${output}

    ${output}=  Run Keyword Unless  ${certs}  Run  ${vic-machine} create --debug 1 --name=%{VCH-NAME} --target=%{TEST_URL}%{TEST_DATACENTER} --thumbprint=%{TEST_THUMBPRINT} --user=%{TEST_USERNAME} --image-store=%{TEST_DATASTORE} --appliance-iso=${appliance-iso} --bootstrap-iso=${bootstrap-iso} --password=%{TEST_PASSWORD} --force=true --bridge-network=%{BRIDGE_NETWORK} --public-network=%{PUBLIC_NETWORK} --compute-resource=%{TEST_RESOURCE} --timeout %{TEST_TIMEOUT} --volume-store=%{TEST_DATASTORE}/test:${vol} --no-tlsverify
    Run Keyword Unless  ${certs}  Should Contain  ${output}  Installer completed successfully
    [Return]  ${output}

Run Secret VIC Machine Delete Command
    [Tags]  secret
    [Arguments]  ${vch-name}
    ${rc}  ${output}=  Run And Return Rc And Output  bin/vic-machine-linux delete --name=${vch-name} --target=%{TEST_URL}%{TEST_DATACENTER} --user=%{TEST_USERNAME} --password=%{TEST_PASSWORD} --force=true --compute-resource=%{TEST_RESOURCE} --timeout %{TEST_TIMEOUT}
    [Return]  ${rc}  ${output}

Run Secret VIC Machine Inspect Command
    [Tags]  secret
    [Arguments]  ${name}
    ${rc}  ${output}=  Run And Return Rc And Output  bin/vic-machine-linux inspect --name=${name} --target=%{TEST_URL}%{TEST_DATACENTER} --user=%{TEST_USERNAME} --password=%{TEST_PASSWORD} --thumbprint=%{TEST_THUMBPRINT}
    [Return]  ${rc}  ${output}

Run VIC Machine Delete Command
    ${rc}  ${output}=  Run Secret VIC Machine Delete Command  %{VCH-NAME}
    Wait Until Keyword Succeeds  6x  5s  Check Delete Success  %{VCH-NAME}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  Completed successfully
    ${output}=  Run  rm -rf %{VCH-NAME}
    [Return]  ${output}

Run VIC Machine Inspect Command
    ${rc}  ${output}=  Run Secret VIC Machine Inspect Command  %{VCH-NAME}
    Get Docker Params  ${output}  ${true}

Gather Logs From Test Server
    [Tags]  secret
    Run Keyword And Continue On Failure  Run  zip %{VCH-NAME}-certs -r %{VCH-NAME}
    ${out}=  Run  curl -k -D vic-admin-cookies -Fusername=%{TEST_USERNAME} -Fpassword=%{TEST_PASSWORD} %{VIC-ADMIN}/authentication
    Log  ${out}
    ${out}=  Run  curl -k -b vic-admin-cookies %{VIC-ADMIN}/container-logs.zip -o ${SUITE NAME}-%{VCH-NAME}-container-logs.zip
    Log  ${out}
    Remove File  vic-admin-cookies
    ${out}=  Run  govc datastore.download %{VCH-NAME}/vmware.log %{VCH-NAME}-vmware.log
    Should Contain  ${out}  OK
    Run Keyword If  '%{HOST_TYPE}' == 'ESXi'  Run  govc logs -log=vmkernel -n=10000 > vmkernel.log

Check For The Proper Log Files
    [Arguments]  ${container}
    # Ensure container logs are correctly being gathered for debugging purposes
    ${rc}  ${output}=  Run And Return Rc and Output  curl -sk %{VIC-ADMIN}/authentication -XPOST -F username=%{TEST_USERNAME} -F password=%{TEST_PASSWORD} -D /tmp/cookies-%{VCH-NAME}
    Should Be Equal As Integers  ${rc}  0
    ${rc}  ${output}=  Run And Return Rc and Output  curl -sk %{VIC-ADMIN}/container-logs.tar.gz -b /tmp/cookies-%{VCH-NAME} | tar tvzf -
    Should Be Equal As Integers  ${rc}  0
    Log  ${output}
    Should Contain  ${output}  ${container}/output.log
    Should Contain  ${output}  ${container}/vmware.log
    Should Contain  ${output}  ${container}/tether.debug

Scrape Logs For the Password
    [Tags]  secret
    ${rc}=  Run And Return Rc  curl -sk %{VIC-ADMIN}/authentication -XPOST -F username=%{TEST_USERNAME} -F password=%{TEST_PASSWORD} -D /tmp/cookies-%{VCH-NAME}
    Should Be Equal As Integers  ${rc}  0

    ${rc}=  Run And Return Rc  curl -sk %{VIC-ADMIN}/logs/port-layer.log -b /tmp/cookies-%{VCH-NAME} | grep -q "%{TEST_PASSWORD}"
    Should Be Equal As Integers  ${rc}  1
    ${rc}=  Run And Return Rc  curl -sk %{VIC-ADMIN}/logs/init.log -b /tmp/cookies-%{VCH-NAME} | grep -q "%{TEST_PASSWORD}"
    Should Be Equal As Integers  ${rc}  1
    ${rc}=  Run And Return Rc  curl -sk %{VIC-ADMIN}/logs/docker-personality.log -b /tmp/cookies-%{VCH-NAME} | grep -q "%{TEST_PASSWORD}"
    Should Be Equal As Integers  ${rc}  1
    ${rc}=  Run And Return Rc  curl -sk %{VIC-ADMIN}/logs/vicadmin.log -b /tmp/cookies-%{VCH-NAME} | grep -q "%{TEST_PASSWORD}"
    Should Be Equal As Integers  ${rc}  1

    Remove File  /tmp/cookies-%{VCH-NAME}

Cleanup VIC Appliance On Test Server
    Log To Console  Gathering logs from the test server %{VCH-NAME}
    Gather Logs From Test Server
    Log To Console  Deleting the VCH appliance %{VCH-NAME}
    ${output}=  Run VIC Machine Delete Command
    Run Keyword And Ignore Error  Cleanup VCH Bridge Network  %{VCH-NAME}
    [Return]  ${output}

Cleanup VCH Bridge Network
    [Arguments]  ${name}
    Run Keyword If  '%{HOST_TYPE}' == 'ESXi'  Run  govc host.portgroup.remove ${name}-bridge
    ${out}=  Run Keyword If  '%{HOST_TYPE}' == 'ESXi'  Run  govc host.portgroup.info
    Run Keyword If  '%{HOST_TYPE}' == 'ESXi'  Should Not Contain  ${out}  ${name}-bridge

Cleanup Datastore On Test Server
    ${out}=  Run  govc datastore.ls
    ${items}=  Split To Lines  ${out}
    :FOR  ${item}  IN  @{items}
    \   ${build}=  Split String  ${item}  -
    \   # Skip any item that is not associated with integration tests
    \   Continue For Loop If  '@{build}[0]' != 'VCH'
    \   # Skip any item that is still running
    \   ${state}=  Get State Of Drone Build  @{build}[1]
    \   Continue For Loop If  '${state}' == 'running'
    \   Log To Console  Removing the following item from datastore: ${item}
    \   ${out}=  Run  govc datastore.rm ${item}
    \   Wait Until Keyword Succeeds  6x  5s  Check Delete Success  ${item}

Cleanup Dangling VMs On Test Server
    ${out}=  Run  govc ls vm
    ${vms}=  Split To Lines  ${out}
    :FOR  ${vm}  IN  @{vms}
    \   ${vm}=  Fetch From Right  ${vm}  /
    \   ${build}=  Split String  ${vm}  -
    \   # Skip any VM that is not associated with integration tests
    \   Continue For Loop If  '@{build}[0]' != 'VCH'
    \   # Skip any VM that is still running
    \   ${state}=  Get State Of Drone Build  @{build}[1]
    \   Continue For Loop If  '${state}' == 'running'
    \   ${uuid}=  Run  govc vm.info -json\=true ${vm} | jq -r '.VirtualMachines[0].Config.Uuid'
    \   Log To Console  Destroying dangling VCH: ${vm}
    \   ${rc}  ${output}=  Run Secret VIC Machine Delete Command  ${vm}
    \   Wait Until Keyword Succeeds  6x  5s  Check Delete Success  ${vm}

Cleanup Dangling Networks On Test Server
    ${out}=  Run  govc ls network
    ${nets}=  Split To Lines  ${out}
    :FOR  ${net}  IN  @{nets}
    \   ${net}=  Fetch From Right  ${net}  /
    \   ${build}=  Split String  ${net}  -
    \   # Skip any Network that is not associated with integration tests
    \   Continue For Loop If  '@{build}[0]' != 'VCH'
    \   # Skip any Network that is still running
    \   ${state}=  Get State Of Drone Build  @{build}[1]
    \   Continue For Loop If  '${state}' == 'running'
    \   ${uuid}=  Run  govc host.portgroup.remove ${net}

Cleanup Dangling vSwitches On Test Server
    ${out}=  Run Keyword If  '%{HOST_TYPE}' == 'ESXi'  Run  govc host.vswitch.info | grep VCH
    ${nets}=  Split To Lines  ${out}
    :FOR  ${net}  IN  @{nets}
    \   ${net}=  Fetch From Right  ${net}  ${SPACE}
    \   ${build}=  Split String  ${net}  -
    \   # Skip any vSwitch that is not associated with integration tests
    \   Continue For Loop If  '@{build}[0]' != 'VCH'
    \   # Skip any vSwitch that is still running
    \   ${state}=  Get State Of Drone Build  @{build}[1]
    \   Continue For Loop If  '${state}' == 'running'
    \   ${uuid}=  Run  govc host.vswitch.remove ${net}

Get Scratch Disk From VM Info
    [Arguments]  ${vm}
    ${disks}=  Run  govc vm.info -json ${vm} | jq -r '.VirtualMachines[].Layout.Disk[].DiskFile[]'
    ${disks}=  Split To Lines  ${disks}
    :FOR  ${disk}  IN  @{disks}
    \   ${disk}=  Fetch From Right  ${disk}  ${SPACE}
    \   ${status}=  Run Keyword And Return Status  Should Contain  ${disk}  scratch.vmdk
    \   Return From Keyword If  ${status}  ${disk}

Cleanup Dangling Containers On Test Server
    ${vms}=  Run  govc ls vm
    ${vms}=  Split To Lines  ${vms}
    :FOR  ${vm}  IN  @{vms}
    \   # Ignore VCH's, we only care about containers at this point
    \   ${status}=  Run Keyword And Return Status  Should Contain  ${vm}  VCH
    \   Continue For Loop If  ${status}
    \   ${disk}=  Get Scratch Disk From VM Info  ${vm}
    \   ${vch}=  Fetch From Left  ${disk}  /
    \   ${vch}=  Split String  ${vch}  -
    \   # Skip any VM that is not associated with integration tests
    \   Continue For Loop If  '@{vch}[0]' != 'VCH'
    \   ${state}=  Get State Of Drone Build  @{vch}[1]
    \   # Skip any VM that is still running
    \   Continue For Loop If  '${state}' == 'running'
    \   # Destroy the VM and remove it from datastore because it is a dangling container
    \   Log To Console  Cleaning up dangling container: ${vm}
    \   ${out}=  Run  govc vm.destroy ${vm}
    \   ${name}=  Fetch From Right  ${vm}  /
    \   ${out}=  Run  govc datastore.rm ${name}
    \   Wait Until Keyword Succeeds  6x  5s  Check Delete Success  ${name}

# VCH upgrade helpers
Install VIC with version to Test Server
    [Arguments]  ${version}=7315  ${insecureregistry}=
    Log To Console  \nDownloading vic ${version} from bintray...
    ${rc}  ${output}=  Run And Return Rc And Output  wget https://bintray.com/vmware/vic-repo/download_file?file_path=vic_${version}.tar.gz -O vic.tar.gz
    ${rc}  ${output}=  Run And Return Rc And Output  tar zxvf vic.tar.gz
    Set Environment Variable  TEST_TIMEOUT  20m0s
    Install VIC Appliance To Test Server  vic-machine=./vic/vic-machine-linux  appliance-iso=./vic/appliance.iso  bootstrap-iso=./vic/bootstrap.iso  certs=${false}  vol=default ${insecureregistry}
    Set Environment Variable  VIC-ADMIN  %{VCH-IP}:2378
    Set Environment Variable  INITIAL-VERSION  ${version}

Clean up VIC Appliance And Local Binary
    Cleanup VIC Appliance On Test Server
    Run  rm -rf vic.tar.gz vic

Upgrade
    Log To Console  \nUpgrading VCH...
    ${rc}  ${output}=  Run And Return Rc And Output  bin/vic-machine-linux upgrade --debug 1 --name=%{VCH-NAME} --target=%{TEST_URL} --user=%{TEST_USERNAME} --password=%{TEST_PASSWORD} --force=true --compute-resource=%{TEST_RESOURCE} --timeout %{TEST_TIMEOUT}
    Should Contain  ${output}  Completed successfully
    Should Not Contain  ${output}  Rolling back upgrade
    Should Be Equal As Integers  ${rc}  0

Check Upgraded Version
    ${rc}  ${output}=  Run And Return Rc And Output  bin/vic-machine-linux version
    @{vers}=  Split String  ${output}
    ${rc}  ${output}=  Run And Return Rc And Output  bin/vic-machine-linux inspect --name=%{VCH-NAME} --target=%{TEST_URL} --thumbprint=%{TEST_THUMBPRINT} --user=%{TEST_USERNAME} --password=%{TEST_PASSWORD} --compute-resource=%{TEST_RESOURCE}
    Should Contain  ${output}  Completed successfully
    Should Contain  ${output}  @{vers}[2]
    Should Not Contain  ${output}  %{INITIAL-VERSION}
    Should Be Equal As Integers  ${rc}  0
    Log  ${output}
    Get Docker Params  ${output}  ${true}

Check Original Version
    ${rc}  ${output}=  Run And Return Rc And Output  bin/vic-machine-linux version
    @{vers}=  Split String  ${output}
    ${rc}  ${output}=  Run And Return Rc And Output  bin/vic-machine-linux inspect --name=%{VCH-NAME} --target=%{TEST_URL} --thumbprint=%{TEST_THUMBPRINT} --user=%{TEST_USERNAME} --password=%{TEST_PASSWORD} --compute-resource=%{TEST_RESOURCE}
    Should Contain  ${output}  Completed successfully
    Should Contain  ${output}  @{vers}[2]
    Should Be Equal As Integers  ${rc}  0
    Log  ${output}
    Get Docker Params  ${output}  ${true}

Rollback
     Log To Console  \nTesting rollback...
    ${rc}  ${output}=  Run And Return Rc And Output  bin/vic-machine-linux upgrade --debug 1 --name=%{VCH-NAME} --target=%{TEST_URL} --user=%{TEST_USERNAME} --password=%{TEST_PASSWORD} --force=true --compute-resource=%{TEST_RESOURCE} --timeout %{TEST_TIMEOUT} --rollback
    Should Contain  ${output}  Completed successfully
    Should Be Equal As Integers  ${rc}  0

Enable VCH SSH
    [Arguments]  ${vic-machine}=bin/vic-machine-linux  ${rootpw}=%{TEST_PASSWORD}  ${target}=%{TEST_URL}  ${password}=%{TEST_PASSWORD}  ${thumbprint}=%{TEST_THUMBPRINT}  ${name}=%{VCH-NAME}  ${user}=%{TEST_USERNAME}  ${resource}=%{TEST_RESOURCE}
    Log To Console  \nEnable SSH on vch...
    ${rc}  ${output}=  Run And Return Rc And Output  ${vic-machine} debug --rootpw ${rootpw} --target ${target} --password ${password} --thumbprint ${thumbprint} --name ${name} --user ${user} --compute-resource ${resource} --enable-ssh
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  Completed successfully
