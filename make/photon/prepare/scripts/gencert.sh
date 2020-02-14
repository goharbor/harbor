#! /bin/bash

# CA key and certificate
openssl req -x509 -nodes -days 365 -newkey rsa:4096 \
        -keyout "harbor_internal_ca.key" \
        -out "harbor_internal_ca.crt" \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware"


# generate proxy key and csr
openssl req -new -newkey rsa:4096 -nodes -sha256 \
        -keyout proxy.key \
        -out proxy.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=proxy"

# Sign proxy
openssl x509 -req -days 365 -sha256 -in proxy.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -out proxy.crt


# generate core key and csr
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout core.key \
        -out core.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=core"

# Sign core csr with CA certificate and key
openssl x509 -req -days 365 -sha256 -in core.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -out core.crt


# job_service key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout job_service.key \
        -out job_service.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=jobservice"

# sign job_service csr with CA certificate and key
openssl x509 -req -days 365 -sha256 -in job_service.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -out job_service.crt

# generate registry key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout registry.key \
        -out registry.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=registry"

# sign registry csr with CA certificate and key
openssl x509 -req -days 365 -sha256 -in registry.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -out registry.crt

# generate registryctl key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout registryctl.key \
        -out registryctl.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=registryctl"

# sign registryctl csr with CA certificate and key
openssl x509 -req -days 365 -sha256 -in registryctl.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -out registryctl.crt



# generate clair_adapter key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout clair_adapter.key \
        -out clair_adapter.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=clair_adapter"

# sign clair_adapter csr with CA certificate and key
openssl x509 -req -days 365 -sha256 -in clair_adapter.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -out clair_adapter.crt


# generate clair key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout clair.key \
        -out clair.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=clair"

# sign clair csr with CA certificate and key
openssl x509 -req -days 365 -sha256 -in clair.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -out clair.crt


# generate notary_signer key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout notary_signer.key \
        -out notary_signer.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=notary_signer"

# sign notary_signer csr with CA certificate and key
openssl x509 -req -days 365 -sha256 -in notary_signer.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -out notary_signer.crt



# generate notary_server key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout notary_server.key \
        -out notary_server.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=notary_server"

# sign notary_server csr with CA certificate and key
openssl x509 -req -days 365 -sha256 -in notary_server.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -out notary_server.crt


# generate chartmuseum key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout chartmuseum.key \
        -out chartmuseum.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=chartmuseum"

# sign chartmuseum csr with CA certificate and key
openssl x509 -req -days 365 -sha256 -in chartmuseum.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -out chartmuseum.crt



# generate harbor_db key
openssl req -new \
        -newkey rsa:4096 -nodes -sha256 -keyout harbor_db.key \
        -out harbor_db.csr \
        -subj "/C=CN/ST=Beijing/L=Beijing/O=VMware/CN=harbor_db"

# sign harbor_db csr with CA certificate and key
openssl x509 -req -days 365 -sha256 -in harbor_db.csr -CA harbor_internal_ca.crt -CAkey harbor_internal_ca.key -CAcreateserial -out harbor_db.crt
