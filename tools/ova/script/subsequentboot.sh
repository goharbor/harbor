#!/bin/bash
set -e
echo "======================= $(date)====================="

export PATH=$PATH:/usr/local/bin

base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $base_dir/common.sh

#configure SSH
configSSH

echo "Adding rules to iptables..."
addIptableRules

#Stop Harbor
echo "Shutting down Harbor..."
down || true

#Garbage collection
value=$(ovfenv -k gc_enabled)
if [ "$value" = "true" ]
then
	echo "GC enabled, starting garbage collection..."
	#If the registry contains no images, the gc will fail.
	#So append a true to avoid failure.
	gc 2>&1 >> /var/log/harbor/gc.log || true
else
	echo "GC disabled, skip garbage collection"
fi

#Configure Harbor
echo "Configuring Harbor..."
configure

#Start Harbor
echo "Starting Harbor..."
up

echo "===================================================="