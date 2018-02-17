#!/usr/bin/env bash

#set -e

SECURITY_DIR="/path/to/secrets/directory"
SUBDOMAIN="foo.bar.example.com"

kubectl create namespace common

kubectl -n common create secret tls ${SUBDOMAIN} \
    --cert=${SECURITY_DIR}/ssl-certificates/ssl.crt \
    --key=${SECURITY_DIR}/ssl-certificates/ssl.key
