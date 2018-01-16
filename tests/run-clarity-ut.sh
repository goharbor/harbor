#!/bin/bash

set -e

cd /harbor_src
 
mv /harbor_resources/node_modules ./

npm install -q --no-progress
npm run test > ./npm-ut-test-results