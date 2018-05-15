#!/bin/bash

set -e

cd /harbor_src
 
mv /harbor_resources/node_modules ./

npm install -q --no-progress
npm run lint
npm run lint:lib
npm run build
npm run test > ./npm-ut-test-results