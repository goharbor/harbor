#!/bin/bash
set -e

cd /build_dir
cp -r /portal_src/* .
ls -la

cat ./package.json
npm install

## Build harbor-ui and link it
rm -rf /build_dir/lib/dist
npm run build:lib
chmod -R +xr /build_dir/lib/dist
cd /build_dir/lib/dist
npm link
cd /build_dir
npm link harbor-ui

## Build production
npm run build:prod

## Unlink
npm unlink harbor-ui