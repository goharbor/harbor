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
	
	echo "Resetting DNS and hostname using vami_ovf_process..."
	/opt/vmware/share/vami/vami_ovf_process --setnetwork || true
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
		sed -i -r s%"#?$cfg_key\s*=\s*.*"%"$cfg_key = $cfg_value"% $cfg_file
	fi
}