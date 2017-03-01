#!/bin/bash

cd /clarity-seed
rm -rf dist/*
cp /angular-cli.json /clarity-seed

npm install
ng build

cp /index.html dist/index.html 

