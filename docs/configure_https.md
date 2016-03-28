#Configure Harbor with HTTPS Access 

Because Harbor does not ship with any certificates, it uses HTTP by default to serve registry requests. This makes it relative simple to configure. However, it is highly recommended that security be enabled for any production environment. Harbor has an Nginx instance as a reverse proxy for all services, you can configure Nginx to enable https.

##Get a certificate

Assuming that your registry’s domain is harbordomain.com, and that its DNS record points to the host where you are running Harbor, you first should get a certificate from a CA. The certificate usually contains a .crt file and a .key file, for example, harbordomain.crt and harbordomain.key.

In a test or development environment, you may choose to use a self-signed certificate instead of the one from a CA. The below command generates your own certificate:
```
openssl req \
  -newkey rsa:4096 -nodes -sha256 -keyout harbordomain.key \
  -x509 -days 365 -out harbordomain.crt
```
Be sure to use harbordomain as a CN.  

##Configuration of Nginx
After obtaining the .crt and .key files, change the directory to Deploy/config/nginx in Harbor project.
```
  cd Deploy/config/nginx
```
Create a new directory “cert/” if it does not exist. Then copy harbordomain.crt and harbordomain.key to cert/.

Rename the existing configuration file of Nginx:
```
  mv nginx.conf nginx.conf.bak
```
Copy the template nginx.https.conf as the new configuration file:
```
  cp nginx.https.conf nginx.conf
```
Edit the file nginx.conf and replace two occurrences of harbordomain.com to your own domain name.
```
  server {
    listen 443 ssl;
    server_name harbordomain.com;

…

    server {
      listen 80;
      server_name harbordomain.com;
      rewrite ^/(.*) https://$server_name$1 permanent;
```


Then look for the SSL section to make sure the files of your certificates match the names in the config file.
```
…
    # SSL
    ssl_certificate /etc/nginx/cert/harbordomain.crt;
    ssl_certificate_key /etc/nginx/cert/harbordomain.key;
```
Save your changes in nginx.conf.

##Installation of Harbor
Next, edit the file Deploy/harbor.cfg , update the hostname and the protocol:
```
  #set hostname
  hostname = habordomain.com
  #set ui_url_protocol
  ui_url_protocol = https
```


Generate configuration files for Harbor:
```
  ./prepare
```
If Harbor is already running, stop and remove the existing instance:
```
  docker-compose stop  
  docker-compose rm
```

Finally, restart Harbor:
```
  docker-compose up –d
```

