#!/bin/bash
set -e

cd /clarity-seed
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

npm install
ng build

cp -r ./src/i18n/ dist/

