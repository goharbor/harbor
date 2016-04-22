#Configuring Harbor with HTTPS Access 

Because Harbor does not ship with any certificates, it uses HTTP by default to serve registry requests. This makes it relatively simple to configure. However, it is highly recommended that security be enabled for any production environment. Harbor has an Nginx instance as a reverse proxy for all services, you can configure Nginx to enable https.

##Getting a certificate

Assuming that your registry's **hostname** is **reg.yourdomain.com**, and that its DNS record points to the host where you are running Harbor. You first should get a certificate from a CA. The certificate usually contains a .crt file and a .key file, for example, **yourdomain.com.crt** and **yourdomain.com.key**.

In a test or development environment, you may choose to use a self-signed certificate instead of the one from a CA. The below commands generate your own certificate:

1) Create your own CA certificate:
```
  openssl req \
    -newkey rsa:4096 -nodes -sha256 -keyout ca.key \
    -x509 -days 365 -out ca.crt
```
2) Generate a Certificate Signing Request, be sure to use **reg.yourdomain.com** as the CN (Common Name):
```
  openssl req \
    -newkey rsa:4096 -nodes -sha256 -keyout yourdomain.com.key \
    -out yourdomain.com.csr
```
3) Generate the certificate of your registry host:

You need to configure openssl first. On Ubuntu, the config file locates at **/etc/ssl/openssl.cnf**. Refer to openssl document for more information. The default CA directory of openssl is called demoCA. Let's create necessary directories and files:
```
  mkdir demoCA
  cd demoCA
  touch index.txt
  echo '01' > serial
  cd ..
 ```
Then run this command to generate the certificate of your registry host:
```
  openssl ca -in yourdomain.com.csr -out yourdomain.com.crt -cert ca.crt -keyfile ca.key -outdir .
```

##Configuration of Nginx
After obtaining the **yourdomain.com.crt** and **yourdomain.com.key** files, change the directory to Deploy/config/nginx in Harbor project.
```
  cd Deploy/config/nginx
```
Create a new directory cert/, if it does not exist. Then copy **yourdomain.com.crt** and **yourdomain.com.key** to cert/, e.g. :
```
  cp yourdomain.com.crt cert/
  cp yourdomain.com.key cert/ 
```

Rename the existing configuration file of Nginx:
```
  mv nginx.conf nginx.conf.bak
```
Copy the template **nginx.https.conf** as the new configuration file:
```
  cp nginx.https.conf nginx.conf
```
Edit the file nginx.conf and replace two occurrences of **harbordomain.com** to your own host name, such as reg.yourdomain.com .
```
  server {
    listen 443 ssl;
    server_name harbordomain.com;

    ...
    
  server {
    listen 80;
    server_name harbordomain.com;
    rewrite ^/(.*) https://$server_name$1 permanent;
```
Then look for the SSL section to make sure the files of your certificates match the names in the config file. Do not change the path of the files.
```
    ...
    
    # SSL
    ssl_certificate /etc/nginx/cert/yourdomain.com.crt;
    ssl_certificate_key /etc/nginx/cert/yourdomain.com.key;
```
Save your changes in nginx.conf.

##Installation of Harbor
Next, edit the file Deploy/harbor.cfg , update the hostname and the protocol:
```
  #set hostname
  hostname = reg.yourdomain.com
  #set ui_url_protocol
  ui_url_protocol = https
```

Generate configuration files for Harbor:
```
./prepare
```
If Harbor is already running, stop and remove the existing instance. Your image data remain in the file system
```
  docker-compose stop
  docker-compose rm
```
Finally, restart Harbor:
```
  docker-compose up -d
```
After setting up HTTPS for Harbor, you can verify it by the follow steps:

1. Open a browser and enter the address: https://reg.yourdomain.com . It should display the user interface of Harbor.

2. On a machine with Docker daemon, make sure the option "-insecure-registry" does not present, run any docker command to verify the setup, e.g. 
```
  docker login reg.yourdomain.com
```
##Troubleshooting
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
    