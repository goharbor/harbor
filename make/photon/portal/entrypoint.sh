#!/bin/bash
set -e

cd /build_dir
cp -r /portal_src/* .
ls -la

cat ./package.json
npm install

## Build harbor-ui and link it
rm -rf /build_dir/lib/dist
npm run build_lib
chmod -R +xr /build_dir/lib/dist
npm run link_lib

## Build production
npm run release