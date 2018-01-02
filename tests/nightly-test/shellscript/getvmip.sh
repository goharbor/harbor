#!/bin/sh

export GOVC_USERNAME=$2
export GOVC_PASSWORD=$3
export GOVC_INSECURE=1
export GOVC_URL=$1

echo $GOVC_URL
echo $GOVC_PASSWORD

govc vm.info -json $4  | jq -r .VirtualMachines[].Guest.HostName