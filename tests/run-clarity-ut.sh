#!/bin/bash

set -e

cd /harbor_src
 
mv /harbor_resources/node_modules ./

npm install
npm run test > ./npm-ut-test-results


