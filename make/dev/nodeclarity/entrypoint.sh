#!/bin/bash
set -e

cd /harbor_ui
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

cp ./src/package.json .
npm install

./node_modules/.bin/ngc -p tsconfig-aot.json
sed -i 's/* as//g' src/app/shared/gauge/gauge.component.js
./node_modules/.bin/rollup -c rollup-config.js

cp -r ./src/i18n/ dist/
cp ./src/styles.css dist/

cp ./node_modules/clarity-icons/clarity-icons.min.css dist/
cp ./node_modules/mutationobserver-shim/dist/mutationobserver.min.js dist/
cp ./node_modules/@webcomponents/custom-elements/custom-elements.min.js dist/
cp ./node_modules/clarity-icons/clarity-icons.min.js dist/
cp ./node_modules/clarity-ui/clarity-ui.min.css dist/
cp -r ./node_modules/clarity-icons/shapes/ dist/
