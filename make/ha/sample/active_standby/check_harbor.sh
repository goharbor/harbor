#!/bin/bash
http_code = `curl -s -o /dev/null -w "%{http_code}" 127.0.0.1`
if [ $http_code == 200 ] || [ $http_code == 301 ] ; then
    exit 0
else
    exit 1
fi
