#!/bin/bash
set -e

attrs=( 
	harbor_admin_password
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
	db_password
	verify_remote_cert
	self_registration
	)
	
cert=/data/cert/server.crt
key=/data/cert/server.key
csr=/data/cert/server.csr
ca_cert=/data/cert/ca.crt
ca_key=/data/cert/ca.key
ext=/data/cert/extfile.cnf

hostname=""

base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../" && pwd )"

isFQDN=true
flag=$base_dir/cert_gen_type

#The location of harbor.cfg
cfg=$base_dir/harbor/harbor.cfg

#Format cert and key files
function format {
	file=$1
	head=$(sed -rn 's/(-+[A-Za-z ]*-+)([^-]*)(-+[A-Za-z ]*-+)/\1/p' $file)
	body=$(sed -rn 's/(-+[A-Za-z ]*-+)([^-]*)(-+[A-Za-z ]*-+)/\2/p' $file)
	tail=$(sed -rn 's/(-+[A-Za-z ]*-+)([^-]*)(-+[A-Za-z ]*-+)/\3/p' $file)
	echo $head > $file
	echo $body | sed  's/\s\+/\n/g' >> $file
	echo $tail >> $file
}

function genCert {
	if [ ! -e $ca_cert ] || [ ! -e $ca_key ]
	then
		openssl req -newkey rsa:4096 -nodes -sha256 -keyout $ca_key \
			-x509 -days 365 -out $ca_cert -subj \
			"/C=US/ST=California/L=Palo Alto/O=VMware/OU=CA/CN=CA"
	fi
	openssl req -newkey rsa:4096 -nodes -sha256 -keyout $key \
		-out $csr -subj \
		"/C=US/ST=California/L=Palo Alto/O=VMware/OU=Harbor/CN=$hostname"
	if [ "$isFQDN" = false ]
	then
		echo "Add subjectAltName = IP: $hostname to certificate"
		echo subjectAltName = IP:$hostname > $ext
		#openssl x509 -req -days 365 -in $csr -signkey $key -extfile $ext -out $cert
		openssl x509 -req -days 365 -in $csr -CA $ca_cert -CAkey $ca_key -CAcreateserial -extfile $ext -out $cert
	else
		#openssl x509 -req -days 365 -in $csr -signkey $key -out $cert
		openssl x509 -req -days 365 -in $csr -CA $ca_cert -CAkey $ca_key -CAcreateserial -out $cert
	fi
	echo "self-signed" > $flag
}

function secure {
	echo "Read attribute using ovfenv: [ ssl_cert ]"
	ssl_cert=$(ovfenv -k ssl_cert)
	echo "Read attribute using ovfenv: [ ssl_cert_key ]"
	ssl_cert_key=$(ovfenv -k ssl_cert_key)
	if [ -n "$ssl_cert" ] && [ -n "$ssl_cert_key" ]
	then
		echo "ssl_cert and ssl_cert_key are both set, using customized certificate"
		echo $ssl_cert > $cert
		format $cert
		echo $ssl_cert_key > $key
		format $key
		echo "customized" > $flag
		return
	fi
	
	if [ ! -e $cert ] || [ ! -e $key ]
	then
		echo "Certificate or key file does not exist, will generate a self-signed certificate"
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
		
	echo "Use the existing certificate and key file"
}

#Modify hostname
hostname=$(hostname --fqdn) || true
if [ -z "$hostname" ]
then
	isFQDN=false
	hostname=$(ip addr show eth0|grep "inet "|tr -s ' '|cut -d ' ' -f 3|cut -d '/' -f 1)
fi

if [ -n "$hostname" ]
then
	echo "Read hostname/IP: [ hostname/IP - $hostname ]"
	sed -i -r s/"hostname\s*=\s*.*"/"hostname = $hostname"/ $cfg
else
	echo "Failed to get the hostname/IP"
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
sed -i -r s%"#?ui_url_protocol\s*=\s*.*"%"ui_url_protocol = $protocol"% $cfg

if [ $protocol = "https" ]
then
	secure
fi

for attr in "${attrs[@]}"
do
	echo "Read attribute using ovfenv: [ $attr ]"
	value=$(ovfenv -k $attr)
	
	#ldap search password and email password can be null
	if [ -n "$value" ] || [ "$attr" = "ldap_search_pwd" ] \
		|| [ "$attr" = "email_password" ]
	then
		if [ "$attr" = ldap_search_pwd ] \
			|| [ "$attr" = email_password ] \
			|| [ "$attr" = db_password ] \
			|| [ "$attr" = harbor_admin_password ]
		then
			bs=$(echo $value | base64)
			#value={base64}$bs
		fi
		sed -i -r s%"#?$attr\s*=\s*.*"%"$attr = $value"% $cfg
	fi
done