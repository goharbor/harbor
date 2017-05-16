#!/bin/bash
NAME=ldap_server
docker rm -f $NAME 2>/dev/null

docker run --env LDAP_ORGANISATION="Harbor." \
--env LDAP_DOMAIN="example.com" \
--env LDAP_ADMIN_PASSWORD="admin" \
-p 389:389 \
-p 636:636 \
--detach --name $NAME osixia/openldap:1.1.7

sleep 3
docker cp ldap_test.ldif ldap_server:/
docker exec ldap_server ldapadd -x -D "cn=admin,dc=example,dc=com" -w admin -f /ldap_test.ldif -ZZ

