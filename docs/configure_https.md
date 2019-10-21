# Configuring Harbor with HTTPS Access

Because Harbor does not ship with any certificates, it uses HTTP by default to serve registry requests. However, using HTTP is acceptable only in air-gapped test or development environments that do not have a connection to the external internet. Using HTTP in environments that are not air-gapped exposes you to man-in-the-middle attacks. In production environments, always use HTTPS. If you enable Content Trust with Notary, you must use HTTPS. 

Harbor uses an `nginx` instance as a reverse proxy for all services. You use the `prepare` script to configure `nginx` to enable HTTPS.

You can use certificates that are signed by a trusted third-party CA, or  you can use self-signed certificates. The following sections describe how to create a CA, and how to use your CA to sign a server certificate and a client certificate. 

## Getting Certificate Authority

```
  openssl genrsa -out ca.key 4096
```
```
  openssl req -x509 -new -nodes -sha512 -days 3650 \
    -subj "/C=TW/ST=Taipei/L=Taipei/O=example/OU=Personal/CN=yourdomain.com" \
    -key ca.key \
    -out ca.crt
```

## Getting Server Certificate

Assuming that your registry's **hostname** is **yourdomain.com**, and that its DNS record points to the host where you are running Harbor. In production environment, you first should get a certificate from a CA. In a test or development environment, you can use your own CA. The certificate usually contains a .crt file and a .key file, for example, **yourdomain.com.crt** and **yourdomain.com.key**.



**1) Create your own Private Key:**

```
  openssl genrsa -out yourdomain.com.key 4096
```

**2) Generate a Certificate Signing Request:**

If you use FQDN like **yourdomain.com** to connect your registry host, then you must use **yourdomain.com** as CN (Common Name).

```
  openssl req -sha512 -new \
    -subj "/C=TW/ST=Taipei/L=Taipei/O=example/OU=Personal/CN=yourdomain.com" \
    -key yourdomain.com.key \
    -out yourdomain.com.csr 
```

**3) Generate the certificate of your registry host:**

Whether you're using FQDN like **yourdomain.com** or IP to connect your registry host, run this command to generate the certificate of your registry host which comply with Subject Alternative Name (SAN) and x509 v3 extension requirement:

**v3.ext**

```
cat > v3.ext <<-EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth 
subjectAltName = @alt_names

[alt_names]
DNS.1=yourdomain.com
DNS.2=yourdomain
DNS.3=hostname
EOF
```

```

  openssl x509 -req -sha512 -days 3650 \
    -extfile v3.ext \
    -CA ca.crt -CAkey ca.key -CAcreateserial \
    -in yourdomain.com.csr \
    -out yourdomain.com.crt
```

## Configuration and Installation

**1) Configure Server Certificate and Key for Harbor**

After obtaining the **yourdomain.com.crt** and **yourdomain.com.key** files, 
you can put them into directory such as ```/root/cert/```:

```
  cp yourdomain.com.crt /data/cert/
  cp yourdomain.com.key /data/cert/ 
```

**2) Configure Server Certificate, Key and CA for Docker**

The Docker daemon interprets ```.crt``` files as CA certificates and ```.cert``` files as client certificates. 

Convert server ```yourdomain.com.crt``` to ```yourdomain.com.cert```:

```
openssl x509 -inform PEM -in yourdomain.com.crt -out yourdomain.com.cert
```
Delpoy ```yourdomain.com.cert```, ```yourdomain.com.key```, and ```ca.crt``` for Docker:

```
  cp yourdomain.com.cert /etc/docker/certs.d/yourdomain.com/
  cp yourdomain.com.key /etc/docker/certs.d/yourdomain.com/
  cp ca.crt /etc/docker/certs.d/yourdomain.com/
```

The following illustrates a configuration with custom certificates:


```
/etc/docker/certs.d/
    └── yourdomain.com:port   
       ├── yourdomain.com.cert  <-- Server certificate signed by CA
       ├── yourdomain.com.key   <-- Server key signed by CA
       └── ca.crt               <-- Certificate authority that signed the registry certificate
```

Notice that you may need to trust the certificate at OS level. Please refer to the [Troubleshooting](#Troubleshooting) section below.

**3) Configure Harbor**

Edit the file `harbor.yml`, update the hostname and uncomment the https block, and update the attributes `certificate` and `private_key`:

```yaml
#set hostname
hostname: yourdomain.com

http:
  port: 80

https:
  # https port for harbor, default is 443
  port: 443
  # The path of cert and key files for nginx
  certificate: /data/cert/yourdomain.com.crt
  private_key: /data/cert/yourdomain.com.key

  ......

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

* Open a browser and enter the address: https://yourdomain.com. It should display the user interface of Harbor. 

* Notice that some browser may still shows the warning regarding Certificate Authority (CA) unknown for security reason even though we signed certificates by self-signed CA and deploy the CA to the place mentioned above. It is because self-signed CA essentially is not a trusted third-party CA. You can import the CA to the browser on your own to solve the warning.

* On a machine with Docker daemon, make sure the option "-insecure-registry" for https://yourdomain.com is not present. 

* If you mapped nginx port 443 to another port, then you should instead create the directory ```/etc/docker/certs.d/yourdomain.com:port``` (or your registry host IP:port). Then run any docker command to verify the setup, e.g.


```
  docker login yourdomain.com
```
If you've mapped nginx 443 port to another, you need to add the port to login, like below:

```
  docker login yourdomain.com:port
```


## Troubleshooting
1. You may get an intermediate certificate from a certificate issuer. In this case, you should merge the intermediate certificate with your own certificate to create a certificate bundle. You can achieve this by the below command:  

    ```
    cat intermediate-certificate.pem >> yourdomain.com.crt 
    ```
2. On some systems where docker daemon runs, you may need to trust the certificate at OS level.  
   On Ubuntu, this can be done by below commands:  
   
    ```sh
    cp yourdomain.com.crt /usr/local/share/ca-certificates/yourdomain.com.crt
    update-ca-certificates
    ```  
    
   On Red Hat (CentOS etc), the commands are:  
   
    ```sh
    cp yourdomain.com.crt /etc/pki/ca-trust/source/anchors/yourdomain.com.crt
    update-ca-trust
    ```
