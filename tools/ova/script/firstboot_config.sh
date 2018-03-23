#!/bin/bash
set -e

base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $base_dir/common.sh

#Configure authentication mode 
echo "Read attribute using ovfenv: [ auth_mode ]"
auth_mode=$(ovfenv -k auth_mode)
if [ -n "$auth_mode" ]
then
	configureHarborCfg "auth_mode" "$auth_mode"
fi

#Configure password of Harbor administrator 
echo "Read attribute using ovfenv: [ harbor_admin_password ]"
adm_pwd=$(ovfenv -k harbor_admin_password)
if [ -n "$adm_pwd" ]
then
	configureHarborCfg "harbor_admin_password" "$adm_pwd"
fi

#Configure password of database 
echo "Read attribute using ovfenv: [ db_password ]"
db_pwd=$(ovfenv -k db_password)
if [ -n "$db_pwd" ]
then
	configureHarborCfg "db_password" "$db_pwd"
fi

#Configure other attrs
configure