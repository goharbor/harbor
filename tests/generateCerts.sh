#!/usr/bin/env bash

# These certs file is only for Harbor testing.
IP='127.0.0.1'
OPENSSLCNF=

for path in /etc/openssl/openssl.cnf /etc/ssl/openssl.cnf /usr/local/etc/openssl/openssl.cnf; do
    if [[ -e ${path} ]]; then
        OPENSSLCNF=${path}
    fi
done
if [[ -z ${OPENSSLCNF} ]]; then
    printf "Could not find openssl.cnf"
    exit 1
fi

# Create CA certificate
openssl req \
    -newkey rsa:4096 -nodes -sha256 -keyout harbor_ca.key \
    -x509 -days 365 -out harbor_ca.crt -subj '/C=CN/ST=PEK/L=Bei Jing/O=VMware/CN=Harbor CA'

# Generate a Certificate Signing Request
openssl req \
    -newkey rsa:4096 -nodes -sha256 -keyout $IP.key \
    -out $IP.csr -subj '/C=CN/ST=PEK/L=Bei Jing/O=VMware/CN=Harbor CA'

# Generate the certificate of local registry host
echo subjectAltName = IP:$IP > extfile.cnf
openssl x509 -req -days 365 -in $IP.csr -CA harbor_ca.crt \
	-CAkey harbor_ca.key -CAcreateserial -extfile extfile.cnf -out $IP.crt	
	
# Copy to harbor default location
mkdir -p /data/cert
cp $IP.crt /data/cert/server.crt
cp $IP.key /data/cert/server.key