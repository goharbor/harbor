#!/bin/bash
set -e

echo "======================= $(date)====================="

export PATH=$PATH:/usr/local/bin

base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $base_dir/common.sh

#Reset root password 
value=$(ovfenv -k root_pwd)
if [ -n "$value" ]
then
	echo "Resetting root password..."
	printf "$value\n$value\n" | passwd root
fi

#configure SSH
configSSH

#echo "Adding rules to iptables..."
addIptableRules

echo "Installing docker compose..."
installDockerCompose

echo "Starting docker service..."
systemctl start docker

echo "Uncompress Harbor offline instaler tar..."
tar -zxvf $base_dir/../harbor-offline-installer*.tgz -C $base_dir/../

echo "Loading images..."
load

echo "Configuring Harbor..."
chmod 600 $base_dir/../harbor/harbor.cfg

#Configure authentication mode 
echo "Read attribute using ovfenv: [ auth_mode ]"
auth_mode=$(ovfenv -k auth_mode)
if [ -n "$auth_mode" ]
then
	sed -i -r s%"#?auth_mode\s*=\s*.*"%"auth_mode = $auth_mode"% $base_dir/../harbor/harbor.cfg
fi

#Configure other attrs
mkdir -p /data/cert/
configure

#Start Harbor
echo "Starting Harbor..."
up

echo "Removing unneeded installation packages..."
rm $base_dir/../harbor-offline-installer*.tgz
rm $base_dir/../harbor/harbor*.tgz

echo "===================================================="