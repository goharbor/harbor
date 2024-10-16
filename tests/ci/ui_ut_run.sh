#!/bin/bash
set -x
set -e

cd ./src/portal
npm install -g -q --no-progress @angular/cli
npm install -g -q --no-progress karma
npm install -q --no-progress

# Run format and check for changes
npm run format

# Check if any files were changed by formatting
if [[ -n $(git status --porcelain) ]]; then
  echo "Formatting issues found. Please run 'npm run format' and commit the changes."
  exit 1
fi

# Lint and run tests
npm run lint
npm run lint:style
npm run test && cd -
