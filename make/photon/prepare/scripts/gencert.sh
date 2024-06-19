#! /bin/bash
set -e

DAYS=365
SUBJECT=/C=CN/ST=Beijing/L=Beijing/O=VMware
CA_KEY="harbor_internal_ca.key"
CA_CRT="harbor_internal_ca.crt"
INTERNAL_TLS_LIST=(
        "proxy" 
        "portal" 
        "core" 
        "job_service" 
        "registry" 
        "registryctl" 
        "trivy_adapter" 
        "harbor_db")

if [ "$#" -eq 0 ]; then
    echo "No arguments provided. Using default values."
    echo DAYS=$DAYS
    echo SUBJECT=$SUBJECT
else
    if [ -n "$1" ]; then
        echo "argument supplied set days to $1"
        DAYS=$1
    fi

    if [ -n "$2" ]; then
        echo "argument supplied set subject to $2"
        SUBJECT=$2
    fi
fi

# CA key and certificate
if [[ ! -f $CA_KEY && ! -f $CA_CRT ]]; then
    openssl req -x509 -nodes -days $DAYS -newkey rsa:4096 \
            -keyout $CA_KEY -out $CA_CRT \
            -subj "${SUBJECT}"
else
    echo "$CA_KEY and $CA_CRT exist, use them to generate certs"
fi

# generate csr, key and cert files
for internal_tls in "${INTERNAL_TLS_LIST[@]}"; do
    openssl req -new -newkey rsa:4096 -nodes -sha256 \
            -keyout ${internal_tls}.key \
            -out ${internal_tls}.csr \
            -subj "${SUBJECT}/CN=${internal_tls}"

    echo subjectAltName = DNS.1:${internal_tls} > extfile.cnf

    openssl x509 -req -days $DAYS -sha256 -in ${internal_tls}.csr \
        -CA $CA_CRT -CAkey $CA_KEY -CAcreateserial \
        -extfile extfile.cnf -out ${internal_tls}.crt
done
