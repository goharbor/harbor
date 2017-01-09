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

function getRegistryVersion {
	registry_version=""
	base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
	registry_version=$(sed -n -e 's|.*library/registry:||p' $base_dir/../harbor/docker-compose.yml)
	if [ -z registry_version ]
	then
		registry_version="latest"
	fi
}

#Garbage collectoin
function gc {
	echo "======================= $(date)====================="

	getRegistryVersion
	
	docker run --name gc --rm --volume /data/registry:/storage \
		--volume $base_dir/../harbor/common/config/registry/:/etc/registry/ \
		registry:$registry_version garbage-collect /etc/registry/config.yml
	
	echo "===================================================="
}

#Add rules to iptables
function addIptableRules {
	iptables -A INPUT -p tcp --dport 5480 -j ACCEPT -w || true
	#iptables -A INPUT -p tcp --dport 5488 -j ACCEPT
	#iptables -A INPUT -p tcp --dport 5489 -j ACCEPT
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

#Configure SSH
function configSSH {
	value=$(ovfenv -k permit_root_login)
	if [ "$value" = "true" ]
	then
		v=yes
	else
		v=no
	fi
	echo "ssh: permit root login - $v"
	sed -i -r s%"^PermitRootLogin .*"%"PermitRootLogin $v"% /etc/ssh/sshd_config
	
	if [ ! -f /etc/ssh/ssh_host_rsa_key ] \
		|| [ ! -f /etc/ssh/ssh_host_ecdsa_key ] \
		|| [ ! -f /etc/ssh/ssh_host_ed25519_key ]
	then
		ssh-keygen -A
	fi
	
	systemctl restart sshd
}

#Configure attr in harbor.cfg
function configureHarborCfg {
	cfg_key=$1
	cfg_value=$2

	basedir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
	cfg_file=$basedir/../harbor/harbor.cfg

	if [ -n "$cfg_key" ]
	then
		cfg_value=$(echo "$cfg_value" | sed -r -e 's%[\/&%]%\\&%g')
		sed -i -r "s%#?$cfg_key\s*=\s*.*%$cfg_key = $cfg_value%" $cfg_file
	fi
}

function pushPhoton {
	set +e
	
	getRegistryVersion
	
	docker run -d --name photon_pusher -v /data/registry:/var/lib/registry -p 5000:5000 registry:$registry_version
	docker tag photon:1.0 127.0.0.1:5000/library/photon:1.0
	sleep 5
	docker push 127.0.0.1:5000/library/photon:1.0
	docker rm -f photon_pusher
	set -e
}