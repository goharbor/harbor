#!/bin/bash

set -e

cd /harbor_src
 
mv /harbor_resources/node_modules ./

npm install -q --no-progress
## Build harbor-ui and link it
npm run build:lib

## Link harbor-ui
chmod -R +xr /harbor_src/lib/dist
cd /harbor_src/lib/dist
npm link
cd /harbor_src
npm link harbor-ui

npm run build
npm run test > ./npm-ut-test-results

rm -rf /harbor_src/node_modules