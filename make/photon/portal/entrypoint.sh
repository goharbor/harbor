#!/bin/bash
set -e

cd /build_dir
cp -r /portal_src/* .
ls -la

# Update
apt-get update
apt-get install -y ruby
ruby -ryaml -rjson -e 'puts JSON.pretty_generate(YAML.load(ARGF))' swagger.yaml>swagger.json

cat ./package.json
npm install

## Build harbor-portal and link it
npm run build_lib
npm run link_lib

## Build production
npm run release
