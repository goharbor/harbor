#! /bin/bash
set -e

if [ -z "$1" ]; then
    echo "No argument supplied set days to 365"
    DAYS=365
else
    echo "No argument supplied set days to $1"
    DAYS=$1
fi

CA_KEY="harbor_internal_ca.key"
CA_CRT="harbor_internal_ca.crt"

# CA key and certificate
if [[ ! -f $CA_KEY && ! -f $CA_CRT ]]; then
openssl req -x509 -nodes -days $DAYS -newkey rsa:4096 \
        -keyout $CA_KEY -out $CA_CRT \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware"
else
    echo "$CA_KEY and $CA_CRT exist, use them to generate certs"
fi

# generate proxy key and csr
openssl req -new -newkey rsa:4096 -nodes -sha256 \
        -keyout proxy.key \
        -out proxy.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=proxy"

# Sign proxy
echo subjectAltName = DNS.1:proxy > extfile.cnf
openssl x509 -req -days $DAYS -sha256 -in proxy.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile extfile.cnf -out proxy.crt

# generate portal key and csr
openssl req -new -newkey rsa:4096 -nodes -sha256 \
        -keyout portal.key \
        -out portal.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=portal"

# Sign portal
echo subjectAltName = DNS.1:portal > extfile.cnf
openssl x509 -req -days $DAYS -sha256 -in portal.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile extfile.cnf -out portal.crt

# generate core key and csr
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout core.key \
        -out core.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=core"

# Sign core csr with CA certificate and key
echo subjectAltName = DNS.1:core > extfile.cnf
openssl x509 -req -days $DAYS -sha256 -in core.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile extfile.cnf -out core.crt


# job_service key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout job_service.key \
        -out job_service.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=jobservice"

# sign job_service csr with CA certificate and key
echo subjectAltName = DNS.1:jobservice > extfile.cnf
openssl x509 -req -days $DAYS -sha256 -in job_service.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile extfile.cnf -out job_service.crt

# generate registry key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout registry.key \
        -out registry.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=registry"

# sign registry csr with CA certificate and key
echo subjectAltName = DNS.1:registry > extfile.cnf
openssl x509 -req -days $DAYS -sha256 -in registry.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile extfile.cnf -out registry.crt

# generate registryctl key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout registryctl.key \
        -out registryctl.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=registryctl"

# sign registryctl csr with CA certificate and key
echo subjectAltName = DNS.1:registryctl > extfile.cnf
openssl x509 -req -days $DAYS -sha256 -in registryctl.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile extfile.cnf -out registryctl.crt


# generate trivy_adapter key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout trivy_adapter.key \
        -out trivy_adapter.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=trivy-adapter"

# sign trivy_adapter csr with CA certificate and key
echo subjectAltName = DNS.1:trivy-adapter > extfile.cnf
openssl x509 -req -days $DAYS -sha256 -in trivy_adapter.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile extfile.cnf -out trivy_adapter.crt


# generate notary_signer key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout notary_signer.key \
        -out notary_signer.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=notary-signer"

# sign notary_signer csr with CA certificate and key
echo subjectAltName = DNS.1:notary-signer > extfile.cnf
openssl x509 -req -days $DAYS -sha256 -in notary_signer.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile extfile.cnf -out notary_signer.crt

# generate notary_server key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout notary_server.key \
        -out notary_server.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=notary-server"

# sign notary_server csr with CA certificate and key
echo subjectAltName = DNS.1:notary-server > extfile.cnf
openssl x509 -req -days $DAYS -sha256 -in notary_server.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile extfile.cnf -out notary_server.crt


# generate chartmuseum key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout chartmuseum.key \
        -out chartmuseum.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=chartmuseum"

# sign chartmuseum csr with CA certificate and key
echo subjectAltName = DNS.1:chartmuseum > extfile.cnf
openssl x509 -req -days $DAYS -sha256 -in chartmuseum.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile extfile.cnf -out chartmuseum.crt


# generate harbor_db key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout harbor_db.key \
        -out harbor_db.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=harbor_db"

# sign harbor_db csr with CA certificate and key
echo subjectAltName = DNS.1:harbor_db > extfile.cnf
openssl x509 -req -days $DAYS -sha256 -in harbor_db.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile extfile.cnf -out harbor_db.crt
