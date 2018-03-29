#!/bin/bash

set -e
#get protocol

#LOG=/var/log/keepalived_check.log
nodeip=$1
nodeaddress="http://${nodeip}"
http_code=`curl -s -o /dev/null -w "%{http_code}" ${nodeaddress}`

if [ $http_code == 200 ] ; then
  protocol="http"
elif [ $http_code == 301 ]
then
  protocol="https"
else
#  echo "`date +"%Y-%m-%d %H:%M:%S"` $1, CHECK_CODE=$http_code" >> $LOG
  exit 1
fi

systeminfo=`curl -k -o - -s ${protocol}://${nodeip}/api/systeminfo`

echo $systeminfo | grep "registry_url"
if [ $? != 0 ] ; then
  exit 1
fi
#TODO need to check Clair, but currently Clair status api is unreachable from LB.
# echo $systeminfo | grep "with_clair" | grep "true"
# if [ $? == 0 ] ; then
# clair is enabled
# do some clair check
# else
# clair is disabled
# fi

#check top api

http_code=`curl -k -s -o /dev/null -w "%{http_code}\n" ${protocol}://${nodeip}/api/repositories/top`
set +e
if [ $http_code == 200 ] ; then
  exit 0
else
  exit 1
fi
