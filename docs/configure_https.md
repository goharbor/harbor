# Configuring Harbor with HTTPS Access

Because Harbor does not ship with any certificates, it uses HTTP by default to serve registry requests.  However, it is highly recommended that security be enabled for any production environment. Harbor has an Nginx instance as a reverse proxy for all services, you can use the prepare script to configure Nginx to enable https.

In a test or development environment, you may choose to use a self-signed certificate instead of the one from a trusted third-party CA. The followings will show you how to create your own CA, and use your CA to sign a server certificate and a client certificate. 

## Getting Certificate Authority

```
  openssl genrsa -out ca.key 4096
```
```
  openssl req -x509 -new -nodes -sha512 -days 3650 \
    -subj "/C=TW/ST=Taipei/L=Taipei/O=example/OU=Personal/CN=a02test12" \
    -key ca.key \
    -out ca.crt
```

## Getting Server Certificate

Assuming that your registry's **hostname** is **reg.yourdomain.com**, and that its DNS record points to the host where you are running Harbor. In production environment, you first should get a certificate from a CA. In a test or development environment, you can use your own CA. The certificate usually contains a .crt file and a .key file, for example, **yourdomain.com.crt** and **yourdomain.com.key**.



**1) Create your own Private Key:**

```
  openssl genrsa -out a02test12.key 4096
```

**2) Generate a Certificate Signing Request:**

If you use FQDN like **reg.yourdomain.com** to connect your registry host, then you must use **reg.yourdomain.com** as CN (Common Name).

```
  openssl req -sha512 -new \
    -subj "/C=TW/ST=Taipei/L=Taipei/O=example/OU=Personal/CN=a02test12" \
    -key a02test12.key \
    -out a02test12.csr 
```

**3) Generate the certificate of your registry host:**

Whether you're using FQDN like **reg.yourdomain.com** or IP to connect your registry host, run this command to generate the certificate of your registry host which supports x509 v3 extenstions:

**v3.ext**

```
cat > v3.ext <<-EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth 
subjectAltName = @alt_names

[alt_names]
DNS.1 = a02test12
EOF
```

```

  openssl x509 -req -sha512 -days 3650 \
    -extfile v3.ext \
    -CA ca.crt -CAkey ca.key -CAcreateserial \
    -in a02test12.csr \
    -out a02test12.crt
```

## Getting Client Certificate

Use OpenSSL’s genrsa and req commands to first generate an RSA key ```client.key``` and then use the key to create the certificate ```client.crt```.

```
  openssl genrsa -out client.key 4096
```

```
  openssl req -sha512 -new \
    -subj "/C=TW/ST=Taipei/L=Taipei/O=example/OU=Personal/CN=a02test12" \
    -key client.key \
    -out client.csr
```
```
  echo extendedKeyUsage = clientAuth >> extfile.cnf
```
```
  openssl x509 -req -days 3650 \
    -in client.csr \
    -CA ca.crt -CAkey ca.key -CAcreateserial\
    -extfile extfile.cnf \
    -out client.cert
```

## Configuration and Installation

**1) Configure Server Certificate and Key**

After obtaining the **yourdomain.com.crt** and **yourdomain.com.key** files, 
you can put them into directory such as ```/root/cert/```:

```
  cp a02test12.crt /data/cert/
  cp a02test12.key /data/cert/ 
```

**2) Configure Client Certificate and CA**

The Docker daemon interprets ```.crt``` files as CA certificates and ```.cert``` files as client certificates. 

```
  cp client.cert /etc/docker/certs.d/a02test12/
  cp client.key /etc/docker/certs.d/a02test12/
  cp ca.crt /etc/docker/certs.d/a02test12/
```
Notice that you may need to trust the certificate at OS level. Please refer to the [Troubleshooting](#troubleshooting) section below.

The following illustrates a configuration with custom certificates:


```
/etc/docker/certs.d/
    └── yourdomain.com   
       ├── client.cert	<-- Client certificate signed by CA
       ├── client.key	<-- Client key signed by CA
       └── ca.crt		<-- Certificate authority that signed the registry certificate
```

**3) Configure Harbor**

Edit the file ```harbor.cfg```, update the hostname and the protocol, and update the attributes ```ssl_cert``` and ```ssl_cert_key```:

```
  #set hostname
  hostname = reg.yourdomain.com
  #set ui_url_protocol
  ui_url_protocol = https
  ......
  #The path of cert and key files for nginx, they are applied only the protocol is set to https 
  ssl_cert = /data/cert/yourdomain.com.crt
  ssl_cert_key = /data/cert/yourdomain.com.key
```

Generate configuration files for Harbor:

```
  ./prepare
```

If Harbor is already running, stop and remove the existing instance. Your image data remain in the file system

```
  docker-compose down -v
```
Finally, restart Harbor:

```
  docker-compose up -d
```
After setting up HTTPS for Harbor, you can verify it by the following steps:

* Open a browser and enter the address: https://reg.yourdomain.com. It should display the user interface of Harbor. 

* Notice that some browser may still shows the warning regarding Certificate Authority (CA) unknown for security reason even though we signed certificates by self-signed CA and deploy the CA to the place mentioned above. It is because self-signed CA essentially is not a trusted third-party CA. You can import the CA to the browser on your own to solve the warning.

* On a machine with Docker daemon, make sure the option "-insecure-registry" for https://reg.yourdomain.com does not present. 

* If you mapped nginx port 443 to another port, then you should instead create the directory ```/etc/docker/certs.d/reg.yourdomain.com:port``` (or your registry host IP:port). Then run any docker command to verify the setup, e.g.


```
  docker login reg.yourdomain.com
```
If you've mapped nginx 443 port to another, you need to add the port to login, like below:

```
  docker login reg.yourdomain.com:port
```


## Troubleshooting
1. You may get an intermediate certificate from a certificate issuer. In this case, you should merge the intermediate certificate with your own certificate to create a certificate bundle. You can achieve this by the below command:  

    ```
    cat intermediate-certificate.pem >> yourdomain.com.crt 
    ```
2. On some systems where docker daemon runs, you may need to trust the certificate at OS level.  
   On Ubuntu, this can be done by below commands:  
   
    ```sh
    cp youdomain.com.crt /usr/local/share/ca-certificates/reg.yourdomain.com.crt
    update-ca-certificates
    ```  
    
   On Red Hat (CentOS etc), the commands are:  
   
    ```sh
    cp yourdomain.com.crt /etc/pki/ca-trust/source/anchors/reg.yourdomain.com.crt
    update-ca-trust
    ```
    
