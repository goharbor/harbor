#!/bin/bash
set -x
set -e

cd ./src/portal
npm install -g -q --no-progress @angular/cli
npm install -g -q --no-progress karma
npm install -q --no-progress
# check i18n keys first to ensure all translation keys used in code exist
node scripts/find-missing-i18n.js
# check code lint first then run ut test
npm run lint
npm run lint:style
npm run test && cd -
