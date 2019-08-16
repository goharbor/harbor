#!/usr/bin/env bash

if [ -z "$1" ];then echo "$0 <buildnum> [action]";exit 1;fi
BUILDNUM=$1
ACTION=${2:-stop}

for name in nginx harbor-jobservice harbor-portal harbor-core registry registryctl harbor-db redis harbor-log; do
  docker $ACTION $name-build.$BUILDNUM
done