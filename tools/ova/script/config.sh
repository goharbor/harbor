#!/bin/bash
set -e

attrs=( 
	harbor_admin_password
	auth_mode 
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

base_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../" && pwd )"

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

#Modify hostname
ip=$(ip addr show eth0|grep "inet "|tr -s ' '|cut -d ' ' -f 3|cut -d '/' -f 1)
if [ -n "$ip" ]
then
	echo "Read IP address: [ IP - $ip ]"
	sed -i -r s/"hostname = .*"/"hostname = $ip"/ $cfg
else
	echo "Failed to get the IP address"
	exit 1
fi

#Handle http/https
protocol=http
echo "Read attribute using ovfenv: [ ssl_cert ]"
ssl_cert=$(ovfenv -k ssl_cert)
echo "Read attribute using ovfenv: [ ssl_cert_key ]"
ssl_cert_key=$(ovfenv -k ssl_cert_key)
if [ -n "$ssl_cert" ] && [ -n "$ssl_cert_key" ]
then
	echo "ssl_cert and ssl_cert_key are set, using HTTPS protocol"
	protocol=https
	sed -i -r s%"#?ui_url_protocol = .*"%"ui_url_protocol = $protocol"% $cfg
	echo $ssl_cert > /data/server.crt
	format /data/server.crt
	echo $ssl_cert_key > /data/server.key
	format /data/server.key
else
	echo "ssl_cert and ssl_cert_key are not set, using HTTP protocol"
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
		sed -i -r s%"#?$attr = .*"%"$attr = $value"% $cfg
	fi
done