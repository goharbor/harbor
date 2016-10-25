#!/bin/bash

#Shut down Harbor
function down {
	base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
	docker-compose -f $base_dir/../harbor/docker-compose*.yml down
}

#Start Harbor
function up {
	base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
	$base_dir/start_harbor.sh
}

#Configure Harbor
function configure {
	base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
	$base_dir/config.sh
}

#Garbage collectoin
function gc {
	echo "======================= $(date)====================="
	
	#the registry image
	image=$1

	base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

	docker run --name gc --rm --volume /data/registry:/storage \
		--volume $base_dir/../harbor/common/config/registry/:/etc/registry/ \
		$image garbage-collect /etc/registry/config.yml
	
	echo "===================================================="
}

#Add rules to iptables
function addIptableRules {
	iptables -A INPUT -p tcp --dport 5480 -j ACCEPT
	iptables -A INPUT -p tcp --dport 5488 -j ACCEPT
	iptables -A INPUT -p tcp --dport 5489 -j ACCEPT
}

#Install docker-compose
function installDockerCompose {
	base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
	$base_dir/../deps/docker-compose-1.7.1/install.sh
}

#Load images
function load {
	basedir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
	docker load -i $basedir/../harbor/harbor*.tgz
}