# Configuring Harbor with HTTPS Access

Because Harbor does not ship with any certificates, it uses HTTP by default to serve registry requests.  However, it is highly recommended that security be enabled for any production environment. Harbor has an Nginx instance as a reverse proxy for all services, you can use the prepare script to configure Nginx to enable https.

In a test or development environment, you may choose to use a self-signed certificate instead of the one from a trusted third-party CA. The followings will show you how to create your own CA, and use your CA to sign a server certificate and a client certificate. 

In https configure guide below, take an example, assuming that your registry's **hostname** is **yourdomain.com** or the IP of machine is `35.243.92.40`.

## Getting Certificate Authority

```bash
  openssl genrsa -out ca.key 4096
```
```bash
  # use domain
  openssl req -x509 -new -nodes -sha512 -days 3650 \
    -subj "/C=TW/ST=Taipei/L=Taipei/O=example/OU=Personal/CN=yourdomain.com" \
    -key ca.key \
    -out ca.crt
    
  # use IP
  openssl req -x509 -new -nodes -sha512 -days 3650 \
      -subj "/C=CN/ST=Zhejiang/L=Hangzhou/O=Example/OU=Personal/CN=35.243.92.40" \
      -key ca.key \
      -out ca.crt
```

## Getting Server Certificate

Assuming that your registry's **hostname** is **yourdomain.com**, and that its DNS record points to the host where you are running Harbor. In production environment, you first should get a certificate from a CA. In a test or development environment, you can use your own CA. The certificate usually contains a .crt file and a .key file, for example, **yourdomain.com.crt** and **yourdomain.com.key**.

When use IP, assuming it is `35.243.92.40`, the certificate files will be `35.243.92.40.key` and `35.243.92.40.crt`.



**1) Create your own Private Key:**

```bash
  # use domain
  openssl genrsa -out yourdomain.com.key 4096
  # use IP
  openssl genrsa -out 35.243.92.40.key 4096
```

**2) Generate a Certificate Signing Request:**

If you use FQDN like **yourdomain.com** to connect your registry host, then you must use **yourdomain.com** as CN (Common Name), otherwise use the IP like `35.243.92.40`.

```bash
  # use domain
  openssl req -sha512 -new \
    -subj "/C=TW/ST=Taipei/L=Taipei/O=example/OU=Personal/CN=yourdomain.com" \
    -key yourdomain.com.key \
    -out yourdomain.com.csr 
  # use IP
  openssl req -sha512 -new \
      -subj "/C=CN/ST=Zhejiang/L=Hangzhou/O=Example/OU=Personal/CN=35.243.92.40" \
      -key 35.243.92.40.key \
      -out 35.243.92.40.csr
```

**3) Generate the certificate of your registry host:**

Whether you're using FQDN like **yourdomain.com** or IP like `35.243.92.40` to connect your registry host, run this command to generate the certificate of your registry host which comply with Subject Alternative Name (SAN) and x509 v3 extension requirement:

**v3.ext**

```bash
# use domain
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

# use IP
cat > v3.ext <<-EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth 
subjectAltName = @alt_names

[alt_names]
IP.1=35.243.92.40
EOF
```

```bash
  # use domain
  openssl x509 -req -sha512 -days 3650 \
    -extfile v3.ext \
    -CA ca.crt -CAkey ca.key -CAcreateserial \
    -in yourdomain.com.csr \
    -out yourdomain.com.crt
  
  # use IP
  openssl x509 -req -sha512 -days 3650 \
      -extfile v3.ext \
      -CA ca.crt -CAkey ca.key -CAcreateserial \
      -in 35.243.92.40.csr \
      -out 35.243.92.40.crt
```

## Configuration and Installation

**1) Configure Server Certificate and Key for Harbor**

After obtaining the **yourdomain.com.crt** and **yourdomain.com.key** files, 
you can put them into directory such as ```/root/cert/```:

```bash
  # use domain
  sudo mkdir -p /data/cert/
  sudo cp yourdomain.com.crt /data/cert/
  sudo cp yourdomain.com.key /data/cert/ 
  
  # use IP
  sudo mkdir -p /data/cert/
  sudo cp 35.243.92.40.crt /data/cert/
  sudo cp 35.243.92.40.key /data/cert/ 
```

**2) Configure Server Certificate, Key and CA for Docker**

The Docker daemon interprets ```.crt``` files as CA certificates and ```.cert``` files as client certificates. 

Convert server ```yourdomain.com.crt``` to ```yourdomain.com.cert``` or `35.243.92.40.crt` to `35.243.92.40.cert`:

```bash
  # use domain
  openssl x509 -inform PEM -in yourdomain.com.crt -out yourdomain.com.cert
  # use IP
  openssl x509 -inform PEM -in 35.243.92.40.crt -out 35.243.92.40.cert
```
Delpoy ```yourdomain.com.cert```, ```yourdomain.com.key```, or `35.243.92.40.cert`, `35.243.92.40.key`, and ```ca.crt``` for Docker:

```bash
  # use domain
  cp yourdomain.com.cert /etc/docker/certs.d/yourdomain.com/
  cp yourdomain.com.key /etc/docker/certs.d/yourdomain.com/
  cp ca.crt /etc/docker/certs.d/yourdomain.com/
  
  # use IP
  sudo mkdir -p /etc/docker/certs.d/35.243.92.40:443/
  sudo cp 35.243.92.40.cert /etc/docker/certs.d/35.243.92.40:443/
  sudo cp 35.243.92.40.key /etc/docker/certs.d/35.243.92.40:443/
  sudo cp ca.crt /etc/docker/certs.d/35.243.92.40:443/
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

Edit the file ```harbor.cfg```, update the hostname and the protocol, and update the attributes ```ssl_cert``` and ```ssl_cert_key```:

```editorconfig
  #set hostname
  hostname = yourdomain.com:port
  #set ui_url_protocol
  ui_url_protocol = https
  ......
  #The path of cert and key files for nginx, they are applied only the protocol is set to https 
  ssl_cert = /data/cert/yourdomain.com.crt
  ssl_cert_key = /data/cert/yourdomain.com.key
```

Start from Harbor 1.8.0, the configure file is `harbor.yml` as the replacement of `harbor.cfg`.
- set the hostname with domain `yourdomain.com` or IP `34.80.154.130`
- comment `http` and its `port`
- uncomment `https`, and config its `port`, `certificate`, and `private_key` like below

```yaml
# The IP address or hostname to access admin UI and registry service.
# DO NOT use localhost or 127.0.0.1, because Harbor needs to be accessed by external clients.
hostname: 34.80.154.130

# http related config
#http:
  # port for http, default is 80. If https enabled, this port will redirect to https port
#  port: 80

# https related config
https:
  # https port for harbor, default is 443
  port: 443
  # The path of cert and key files for nginx
  certificate: /data/cert/34.80.154.130.crt
  private_key: /data/cert/34.80.154.130.key

# ...
```

Generate configuration files for Harbor:

```bash
  sudo ./prepare
```
if you want to enable notary and clair, it can be like below.
```bash
  sudo ./prepare --with-notary --with-clair 
```

If Harbor is already running, stop and remove the existing instance. Your image data remain in the file system

```bash
  sudo docker-compose down -v
```
Finally, restart Harbor:

```bash
  sudo docker-compose up -d
```
After setting up HTTPS for Harbor, you can verify it by the following steps:

* Open a browser and enter the address: `https://yourdomain.com` or `https://34.80.154.130`. It should display the user interface of Harbor. 

* Notice that some browser may still shows the warning regarding Certificate Authority (CA) unknown for security reason even though we signed certificates by self-signed CA and deploy the CA to the place mentioned above. It is because self-signed CA essentially is not a trusted third-party CA. You can import the CA to the browser on your own to solve the warning.

* On a machine with Docker daemon, make sure the option "-insecure-registry" for https://yourdomain.com does not present. 

* If you mapped nginx port 443 to another port, then you should instead create the directory ```/etc/docker/certs.d/yourdomain.com:port``` (or your registry host IP:port). Then run any docker command to verify the setup, e.g.


```bash
  sudo docker login yourdomain.com
  # or sudo docker login 34.80.154.130
```
If you've mapped nginx 443 port to another, you need to add the port to login, like below:

```bash
  docker login yourdomain.com:port
  # or sudo docker login 34.80.154.130:port
```


##Troubleshooting
1. You may get an intermediate certificate from a certificate issuer. In this case, you should merge the intermediate certificate with your own certificate to create a certificate bundle. You can achieve this by the below command:  

    ```bash
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
3. If you want to use self-signed CA  when remote certificate verify, you can manually add your self-signed CA into the `harbor-core` container, as shows below:

    ```bash
    # for harbor1(35.243.92.40)
    docker cp ca.crt  harbor-core:/etc/ssl/certs/
    # or the certificate of server harbor2(34.80.154.130)
    # docker cp 34.80.154.130.crt  harbor-core:/etc/ssl/certs/
    
    docker restart harbor-core
    ```
