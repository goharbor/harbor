#!/bin/sh

export GOVC_USERNAME=$2
export GOVC_PASSWORD=$3
export GOVC_INSECURE=1
export GOVC_URL=$1
govc vm.destroy $4
