#!/bin/env bash

if [ ! -z "$*" ]; then
 $@ 2>&1 | while read line;do
  echo $(date +"%T") $line
 done
 exit ${PIPESTATUS[0]}
else
 while read line;do
   echo $(date +"%T") $line
  done
  echo ret $?
fi