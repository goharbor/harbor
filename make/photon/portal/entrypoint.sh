#!/bin/bash
set -e

cd /build_dir
cp -r /portal_src/* .
ls -la

cat ./package.json
npm install

## Build harbor-portal and link it
npm run build_lib
npm run link_lib

## Build production
npm run release