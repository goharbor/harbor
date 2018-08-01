#!/bin/bash
set -e

cd /harbor_src/ui_ng
rm -rf dist/*

npm_proxy=

while getopts p: option
do
    case "${option}"
    in
		p) npm_proxy=${OPTARG};;
    esac
done

if [ ! -z "$npm_proxy" -a "$npm_proxy" != " " ]; then
	npm config set proxy $npm_proxy
fi

#Check if node_modules directory existing
if [ ! -d "./node_modules" ]; then
  mv /harbor_resources/node_modules ./
fi

cat ./package.json
npm install

## Build harbor-ui and link it
rm -rf /harbor_src/ui_ng/lib/dist
npm run build:lib
chmod -R +xr /harbor_src/ui_ng/lib/dist
cd /harbor_src/ui_ng/lib/dist
npm link
cd /harbor_src/ui_ng
npm link harbor-ui

## Rollup
./node_modules/.bin/ngc -p tsconfig-aot.json
sed -i 's/* as//g' src/app/shared/gauge/gauge.component.js
./node_modules/.bin/rollup -c rollup-config.js

## Unlink
npm unlink harbor-ui

#Copy built js to the static folder
cp ./dist/build.min.js ../ui/static/

cp -r ./src/i18n/ ../ui/static/
cp ./src/styles.css ../ui/static/
cp -r ./src/images/ ../ui/static/
cp ./src/setting.json ../ui/static/

cp ./node_modules/clarity-icons/clarity-icons.min.css ../ui/static/
cp ./node_modules/mutationobserver-shim/dist/mutationobserver.min.js ../ui/static/
cp ./node_modules/@webcomponents/custom-elements/custom-elements.min.js ../ui/static/
cp ./node_modules/clarity-icons/clarity-icons.min.js ../ui/static/
cp ./node_modules/clarity-ui/clarity-ui.min.css ../ui/static/
cp -r ./node_modules/clarity-icons/shapes/ ../ui/static/

cp ./node_modules/prismjs/themes/prism-solarizedlight.css ../ui/static/
cp ./node_modules/marked/lib/marked.js ../ui/static/
cp ./node_modules/prismjs/prism.js ../ui/static/
cp ./node_modules/prismjs/components/prism-yaml.min.js ../ui/static/