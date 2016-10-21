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
2) Generate a Certificate Signing Request:

If you use FQDN like **reg.yourdomain.com** to connect your registry host, then you must use **reg.yourdomain.com** as CN (Common Name). 
Otherwise, if you use IP address to connect your registry host, CN can be anything like your name and so on:
```
  openssl req \
    -newkey rsa:4096 -nodes -sha256 -keyout yourdomain.com.key \
    -out yourdomain.com.csr
```
3) Generate the certificate of your registry host:

On Ubuntu, the config file of openssl locates at **/etc/ssl/openssl.cnf**. Refer to openssl document for more information. The default CA directory of openssl is called demoCA. Let's create necessary directories and files:
```
  mkdir demoCA
  cd demoCA
  touch index.txt
  echo '01' > serial
  cd ..
 ```
If you're using FQDN like **reg.yourdomain.com** to connect your registry host, then run this command to generate the certificate of your registry host:
```
  openssl ca -in yourdomain.com.csr -out yourdomain.com.crt -cert ca.crt -keyfile ca.key -outdir .
```
If you're using **IP** to connect your registry host, you may instead run the command below:
```
  
  echo subjectAltName = IP:your registry host IP > extfile.cnf

  openssl ca -in yourdomain.com.csr -out yourdomain.com.crt -cert ca.crt -keyfile ca.key -extfile extfile.cnf -outdir .
```
##Configuration of Nginx
After obtaining the **yourdomain.com.crt** and **yourdomain.com.key** files, change the directory to make/config/nginx in Harbor project.
```
  cd make/config/nginx
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
Edit the file nginx.conf and replace two occurrences of **harbordomain.com** to your own host name, such as reg.yourdomain.com . If you use a customized port rather than the default port 443, replace the port "443" in the line "rewrite ^/(.*) https://$server_name:443/$1 permanent;" as well. Please refer to the [installation guide](https://github.com/vmware/harbor/blob/master/docs/installation_guide.md) for other required steps of port customization.  
```
  server {
    listen 443 ssl;
    server_name harbordomain.com;

    ...
    
  server {
    listen 80;
    server_name harbordomain.com;
    rewrite ^/(.*) https://$server_name:443/$1 permanent;
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
Next, edit the file make/harbor.cfg , update the hostname and the protocol:
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
After setting up HTTPS for Harbor, you can verify it by the following steps:

1. Open a browser and enter the address: https://reg.yourdomain.com . It should display the user interface of Harbor.

2. On a machine with Docker daemon, make sure the option "-insecure-registry" does not present, and you must copy ca.crt generated in the above step to /etc/docker/certs.d/yourdomain.com(or your registry host IP), if the directory does not exist, create it.
If you mapped nginx port 443 to another port, then you should instead create the directory /etc/docker/certs.d/yourdomain.com:port(or your registry host IP:port). Then run any docker command to verify the setup, e.g. 

```
  docker login reg.yourdomain.com
```
If you've mapped nginx 443 port to another, you need to add the port to login, like below:

```
  docker login reg.yourdomain.com:port
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
    
