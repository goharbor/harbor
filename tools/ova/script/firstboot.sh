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

#echo "Adding rules to iptables..."
#addIptableRules

echo "Installing docker compose..."
installDockerCompose

echo "Starting docker service..."
systemctl start docker

echo "Uncompress Harbor offline instaler tar..."
tar -zxvf $base_dir/../harbor-offline-installer*.tgz -C $base_dir/../

echo "Loading images..."
load

#Configure Harbor
echo "Configuring Harbor..."
chmod 600 $base_dir/../harbor/harbor.cfg
configure

#Start Harbor
echo "Starting Harbor..."
up

echo "===================================================="