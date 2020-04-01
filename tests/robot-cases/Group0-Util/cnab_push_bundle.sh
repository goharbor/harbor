#!/bin/bash

IP=$1
USER=$2
PWD=$3
TARGET=$4
BUNDLE_FILE=$5
echo $IP

docker login $IP -u $USER -p $PWD
cnab-to-oci fixup  $BUNDLE_FILE --target $TARGET --bundle fixup_bundle.json --auto-update-bundle
cnab-to-oci push  fixup_bundle.json --target $TARGET --auto-update-bundle