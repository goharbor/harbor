#!/usr/bin/env bash
set -e

# Generate new version number by git tag and drone id
base_version="0.5"
version_config="\"version\": \"$base_version.$DRONE_BUILD_NUMBER\""
harbor_root="/drone/src/github.com/vmware/harbor"
ui_lib_path="/src/ui_ng/lib"
npm_token_script_path="/tools/ui_lib/get_npm_token.py"

TOKEN=$($harbor_root$npm_token_script_path)
npm set //registry.npmjs.org/:_authToken $TOKEN

echo "Build harbor-ui lib ..."
cd $harbor_root$ui_lib_path
npm install
npm run build

echo "Update package file for VIC ..."
cd ./dist
# update lib name for VIC
sed -i -e 's/harbor-ui/harbor-ui-vic/1' package.json
# update drone number based version number 
sed -i -e "s/\"version\":[[:space:]]\".*\"/$version_config/g" package.json
npm publish