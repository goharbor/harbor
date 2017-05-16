#!/bin/bash
set -e

attrs=( 
	ldap_url 
	ldap_searchdn 
	ldap_search_pwd 
	ldap_basedn 
	ldap_uid 
	email_server 
	email_server_port 
	email_username 
	email_password 
	email_from 
	email_ssl 
	verify_remote_cert
	self_registration
	)
	
cert_dir=/data/cert
mkdir -p $cert_dir

cert=$cert_dir/server.crt
key=$cert_dir/server.key
csr=$cert_dir/server.csr
ca_cert=$cert_dir/ca.crt
ca_key=$cert_dir/ca.key
ext=$cert_dir/extfile.cnf

ca_download_dir=/data/ca_download
mkdir -p $ca_download_dir
rm -rf $ca_download_dir/*

hostname=""
ip_addr=""

base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../" && pwd )"
source $base_dir/script/common.sh

flag=$base_dir/cert_gen_type

#The location of harbor.cfg
cfg=$base_dir/harbor/harbor.cfg

#Format cert file
function formatCert {
	content=$1
	file=$2
	echo $content | sed -r "s/ /\n/g" | sed -r "/-+$/d" | sed -r "s/^(-+).*/& CERTIFICATE\1/g" > $file
}

#Format key file
function formatKey {
	content=$1
	file=$2
	echo $content | sed -r "s/ /\n/g" | sed -r "/^PRIVATE$/d"| sed -r "/-+$/d" | sed -r "s/^(-+).*/& PRIVATE KEY\1/g" > $file
}

function genCert {
	if [ ! -e $ca_cert ] || [ ! -e $ca_key ]
	then
		openssl req -newkey rsa:4096 -nodes -sha256 -keyout $ca_key \
			-x509 -days 365 -out $ca_cert -subj \
			"/C=US/ST=California/L=Palo Alto/O=VMware, Inc./OU=Harbor/CN=Self-signed by VMware, Inc."
	fi
	openssl req -newkey rsa:4096 -nodes -sha256 -keyout $key \
		-out $csr -subj \
		"/C=US/ST=California/L=Palo Alto/O=VMware/OU=Harbor/CN=$hostname"
	
	echo "Add subjectAltName = IP: $ip_addr to certificate"
	echo subjectAltName = IP:$ip_addr > $ext
	openssl x509 -req -days 365 -in $csr -CA $ca_cert -CAkey $ca_key -CAcreateserial -extfile $ext -out $cert
	
	echo "self-signed" > $flag
	echo "Copy CA certificate to $ca_download_dir"
	cp $ca_cert $ca_download_dir/
}

function secure {
	echo "Read attribute using ovfenv: [ ssl_cert ]"
	ssl_cert=$(ovfenv -k ssl_cert)
	echo "Read attribute using ovfenv: [ ssl_cert_key ]"
	ssl_cert_key=$(ovfenv -k ssl_cert_key)
	if [ -n "$ssl_cert" ] && [ -n "$ssl_cert_key" ]
	then
		echo "ssl_cert and ssl_cert_key are both set, using customized certificate"
		formatCert "$ssl_cert" $cert
		formatKey "$ssl_cert_key" $key
		echo "customized" > $flag
		return
	fi
	
	if [ ! -e $ca_cert ] || [ ! -e $cert ] || [ ! -e $key ]
	then
		echo "CA, Certificate or key file does not exist, will generate a self-signed certificate"
		genCert
		return
	fi
	
	if [ ! -e $flag ] 
	then
		echo "The file which records the way generating certificate does not exist, will generate a new self-signed certificate"
		genCert
		return
	fi
	
	if [ ! $(cat $flag) = "self-signed" ]
	then
		echo "The way generating certificate changed, will generate a new self-signed certificate"
		genCert
		return
	fi
	
	cn=$(openssl x509 -noout -subject -in $cert | sed -n '/^subject/s/^.*CN=//p') || true
	if [ "$hostname" !=  "$cn" ]
	then
		echo "Common name changed: $cn -> $hostname , will generate a new self-signed certificate"
		genCert
		return
	fi
	
	ip_in_cert=$(openssl x509 -noout -text -in $cert | sed -n '/IP Address:/s/.*IP Address://p') || true
	if [ "$ip_addr" !=  "$ip_in_cert" ]
	then
		echo "IP changed: $ip_in_cert -> $ip_addr , will generate a new self-signed certificate"
		genCert
		return
	fi
	
	echo "Use the existing CA, certificate and key file"
	echo "Copy CA certificate to $ca_download_dir"
	cp $ca_cert $ca_download_dir/
}

function detectHostname {
	hostname=$(hostname --fqdn) || true
	if [ -n $hostname ]
	then
		if [ "$hostname" = "localhost.localdom" ]
		then
			hostname=""
			return
		fi
		echo "Get hostname from command 'hostname --fqdn': $hostname"
		return
	fi
}

#Modify hostname
detectHostname
ip_addr=$(ip addr show eth0|grep "inet "|tr -s ' '|cut -d ' ' -f 3|cut -d '/' -f 1)
if [ -z "$hostname" ]
then
	echo "Hostname is null, set it to IP"
	hostname=$ip_addr
fi

if [ -n "$hostname" ]
then
	echo "Hostname: $hostname"
	configureHarborCfg "hostname" "$hostname"
else
	echo "Failed to get the hostname"
	exit 1
fi

#Handle http/https
echo "Read attribute using ovfenv: [ protocol ]"
protocol=$(ovfenv -k protocol)
if [ -z $protocol ]
then
	protocol=https
fi

echo "Protocol: $protocol"
configureHarborCfg ui_url_protocol $protocol

if [ $protocol = "https" ]
then
	secure
fi

for attr in "${attrs[@]}"
do
	echo "Read attribute using ovfenv: [ $attr ]"
	value=$(ovfenv -k $attr)
	
	#if [ "$attr" = ldap_search_pwd ] \
	#	|| [ "$attr" = email_password ]
	#then
	#	bs=$(echo $value | base64)
	#	value={base64}$bs
	#fi
	configureHarborCfg "$attr" "$value"
done