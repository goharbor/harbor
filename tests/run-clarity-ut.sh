#!/bin/bash

set -e

cp -r /harbor_ui/lib/* /harbor_ui

npm install
npm run build
npm run test > lib/npm-ut-test-results


