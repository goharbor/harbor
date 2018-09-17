#!/bin/bash
IP=`ip addr s eth0 |grep "inet "|awk '{print $2}' |awk -F "/" '{print $1}'`
PROTOCOL='https'

#echo $IP
sudo sed "s/reg.mydomain.com/$IP/" -i make/harbor.cfg
sudo sed "s/^ui_url_protocol = .*/ui_url_protocol = $PROTOCOL/g" -i make/harbor.cfg

if [ "$1" = 'LDAP' ]; then
    sudo sed "s/db_auth/ldap_auth/" -i make/harbor.cfg
    sudo sed "s/ldaps:\/\/ldap.mydomain.com/ldap:\/\/$IP/g" -i make/harbor.cfg
    sudo sed "s/#ldap_searchdn = uid=searchuser,ou=people,dc=mydomain,dc=com/ldap_searchdn = cn=admin,dc=example,dc=com/" -i make/harbor.cfg
    sudo sed "s/#ldap_search_pwd = password/ldap_search_pwd = admin/" -i make/harbor.cfg
    sudo sed "s/ldap_basedn = ou=people,dc=mydomain,dc=com/ldap_basedn = dc=example,dc=com/" -i make/harbor.cfg
    sudo sed "s/ldap_uid = uid/ldap_uid = cn/" -i make/harbor.cfg
fi