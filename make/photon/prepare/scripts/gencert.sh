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
cat <<END > proxy.cnf
subjectAltName = @alt_names
[alt_names]
DNS.1 = proxy
DNS.2 = localhost
IP.1 = 127.0.0.1
END
openssl req -new -newkey rsa:4096 -nodes -sha256 \
        -keyout proxy.key \
        -out proxy.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=proxy"

# Sign proxy
openssl x509 -req -days $DAYS -sha256 -in proxy.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile proxy.cnf -out proxy.crt


# generate core key and csr
cat <<END > core.cnf
subjectAltName = @alt_names
[alt_names]
DNS.1 = core
DNS.2 = localhost
IP.1 = 127.0.0.1
END
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout core.key \
        -out core.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=core"

# Sign core csr with CA certificate and key
openssl x509 -req -days $DAYS -sha256 -in core.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile core.cnf -out core.crt


# job_service key
cat <<END > job_service.cnf
subjectAltName = @alt_names
[alt_names]
DNS.1 = jobservice
DNS.2 = localhost
IP.1 = 127.0.0.1
END
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout job_service.key \
        -out job_service.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=jobservice"

# sign job_service csr with CA certificate and key
openssl x509 -req -days $DAYS -sha256 -in job_service.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile job_service.cnf -out job_service.crt

# generate registry key
cat <<END > registry.cnf
subjectAltName = @alt_names
[alt_names]
DNS.1 = registry
DNS.2 = localhost
IP.1 = 127.0.0.1
END
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout registry.key \
        -out registry.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=registry"

# sign registry csr with CA certificate and key
openssl x509 -req -days $DAYS -sha256 -in registry.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile registry.cnf -out registry.crt

# generate registryctl key
cat <<END > registryctl.cnf
subjectAltName = @alt_names
[alt_names]
DNS.1 = registryctl
DNS.2 = localhost
IP.1 = 127.0.0.1
END
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout registryctl.key \
        -out registryctl.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=registryctl"

# sign registryctl csr with CA certificate and key
openssl x509 -req -days $DAYS -sha256 -in registryctl.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile registryctl.cnf -out registryctl.crt



# generate clair_adapter key
cat <<END > clair_adapter.cnf
subjectAltName = @alt_names
[alt_names]
DNS.1 = clair-adapter
DNS.2 = localhost
IP.1 = 127.0.0.1
END
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout clair_adapter.key \
        -out clair_adapter.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=clair-adapter"

# sign clair_adapter csr with CA certificate and key
openssl x509 -req -days $DAYS -sha256 -in clair_adapter.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile clair_adapter.cnf -out clair_adapter.crt


# generate clair key
cat <<END > clair.cnf
subjectAltName = @alt_names
[alt_names]
DNS.1 = clair
DNS.2 = localhost
IP.1 = 127.0.0.1
END
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout clair.key \
        -out clair.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=clair"

# sign clair csr with CA certificate and key
openssl x509 -req -days $DAYS -sha256 -in clair.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile clair.cnf -out clair.crt


# generate trivy_adapter key
cat <<END > trivy_adapter.cnf
subjectAltName = @alt_names
[alt_names]
DNS.1 = trivy-adapter
DNS.2 = localhost
IP.1 = 127.0.0.1
END
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout trivy_adapter.key \
        -out trivy_adapter.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=trivy-adapter"

# sign trivy_adapter csr with CA certificate and key
openssl x509 -req -days $DAYS -sha256 -in trivy_adapter.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile trivy_adapter.cnf -out trivy_adapter.crt


# generate notary_signer key
cat <<END > notary_signer.cnf
subjectAltName = @alt_names
[alt_names]
DNS.1 = notary-signer
DNS.2 = localhost
IP.1 = 127.0.0.1
END
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout notary_signer.key \
        -out notary_signer.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=notary-signer"

# sign notary_signer csr with CA certificate and key
openssl x509 -req -days $DAYS -sha256 -in notary_signer.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile notary_signer.cnf -out notary_signer.crt

# generate notary_server key
cat <<END > notary_server.cnf
subjectAltName = @alt_names
[alt_names]
DNS.1 = notary-server
DNS.2 = localhost
IP.1 = 127.0.0.1
END
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout notary_server.key \
        -out notary_server.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=notary-server"

# sign notary_server csr with CA certificate and key
openssl x509 -req -days $DAYS -sha256 -in notary_server.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile notary_server.cnf -out notary_server.crt


# generate chartmuseum key
cat <<END > chartmuseum.cnf
subjectAltName = @alt_names
[alt_names]
DNS.1 = chartmuseum
DNS.2 = localhost
IP.1 = 127.0.0.1
END
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout chartmuseum.key \
        -out chartmuseum.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=chartmuseum"

# sign chartmuseum csr with CA certificate and key
openssl x509 -req -days $DAYS -sha256 -in chartmuseum.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile chartmuseum.cnf -out chartmuseum.crt


# generate harbor_db key
cat <<END > harbor_db.cnf
subjectAltName = @alt_names
[alt_names]
DNS.1 = harbor_db
DNS.2 = localhost
IP.1 = 127.0.0.1
END
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout harbor_db.key \
        -out harbor_db.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=harbor_db"

# sign harbor_db csr with CA certificate and key
openssl x509 -req -days $DAYS -sha256 -in harbor_db.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -extfile harbor_db.cnf -out harbor_db.crt
