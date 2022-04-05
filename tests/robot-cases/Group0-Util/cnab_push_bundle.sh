#!/bin/bash
set -x

IP=$1
USER=$2
PWD=$3
TARGET=$4
BUNDLE_FILE=$5
REGISTRY=$6
NAMESPACE=$7
IMAGE1=$8
IMAGE2=$9

sed -i "s/registry/$REGISTRY/g" "$BUNDLE_FILE"
sed -i "s/namespace/$NAMESPACE/g" "$BUNDLE_FILE"
sed -i "s/image1/$IMAGE1/g" "$BUNDLE_FILE"
sed -i "s/image2/$IMAGE2/g" "$BUNDLE_FILE"

docker login "$IP" -u "$USER" -p "$PWD"
cnab-to-oci fixup "$BUNDLE_FILE" --target "$TARGET" --bundle fixup_bundle.json --auto-update-bundle
cnab-to-oci push fixup_bundle.json --target "$TARGET" --auto-update-bundle