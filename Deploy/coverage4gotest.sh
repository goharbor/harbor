#!/bin/bash
set -e
echo "mode: set" >>profile.cov
for dir in $(go list ./... | grep -v -E 'vendor|tests')
do
  go test -cover -coverprofile=profile.tmp $dir
  if [ -f profile.tmp ]
  then
      cat profile.tmp | tail -n +2 >> profile.cov
      rm profile.tmp
  fi
done
