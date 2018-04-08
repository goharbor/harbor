#!/bin/sh

psql -h "localhost" -U "postgres" -c 'select 1'
ret_code=$?

if [ $ret_code != 0 ]; then
  exit 1
fi
