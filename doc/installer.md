#Harbor Installation
##Getting harbor

Getting harbor released package file from release directory under harbor source code repository.

###Control Machine Requirements
Currently install harbor need docker-engine, docker compose and python installed on your machine.
docker engine version require 1.8 or higher, 
docker-compose version require 1.6.0 or higher, 
Python verion require 2.7 or higher some Linux distributions (Gentoo, Arch) may not have a Python 2.7 interpreter installed by default. On those systems, you should install one.  
[Install docker engine](https://docs.docker.com/engine/installation/)  
[Install docker compose](https://docs.docker.com/compose/install/)  

##Configure harbor parameters 
For example you get the released package file harbor-o.1.tgz

```
$ tar -xzvf harbor-o.1.tgz
$ cd harbor  
# Make update to the parameters in ./harbor.cfg  
# Simply you just need to change the value of hostname to your machine IP address for example 192.168.1.10
# Generate configuration file,those files are used for harbor related container  
$ ./prepare
Generated configuration file: ./config/ui/env
Generated configuration file: ./config/ui/app.conf
Generated configuration file: ./config/registry/config.yml
Generated configuration file: ./config/db/env
The configuration files are ready, please use docker-compose to start the service.
```


##Build service images
Harbor service buiild as docker images and use docker-compose start the services.
harbor_ui, harbor_mysql, harbor_log are build from docker file and other images like nginx, registry are pull from Docker Hub. Pull image from Docker Hub need Internet access.

```
$ cd harbor
$ sudo docker-compose build
```
##Run harbor services
After all the images build use docker compose start harbor services 
```
$ sudo docker-compose up -d
$ sudo docker ps
# List all the running services
```
Open the Brower access the hostname that you updated in harbor.cfg you will get harbor homepage.  

In case of run harbor on the machine whicn in private network, you should save the docker images.
```
$ sudo ./save_image.sh  
# all the iamges saved to ./images
$ cd ../  
$ tar -cvzf harbor_offline-0.1.tgz harbor
```
harbor_offline-0.1.tgz include all the things used to run harbor services.
Then send harbor_offline-0.1.tgz to destination machine reconfigure the harbor parameters and load images build by previously step.  
Use docker compose start services.
```
$ tar -xzvf harbor_offline-0.1.tgz  
$ cd harbor  
# load images save by excute ./save_image.sh
$ ./load_image.sh
# Make update to the parameters in ./harbor.cfg  
# Simply you just need to change the value of hostname to your machine IP address for example 192.168.1.10 or mydomian.com
# Generate configuration file,those files are used for harbor related container  
$ ./prepare
Generated configuration file: ./config/ui/env
Generated configuration file: ./config/ui/app.conf
Generated configuration file: ./config/registry/config.yml
Generated configuration file: ./config/db/env
The configuration files are ready, please use docker-compose to start the service.
# Build the images and then start the services
$ sudo docker-compose up -d
```

Open the Brower access the hostname that you updated in harbor.cfg you will get harbor homepage.  

### Parameters in harbor.cfg
**hostname**: The endpoint for user to access UI and registry service, for example 192.168.1.10 or exampledomian.com.  
**ui_url_protocol**: The protocol for accessing the UI and token/notification service, by default it is http.User can set it to https if ssl is setup on nginx, for example http or https.  
Email settings for ui to send password resetting email  
* email_server = smtp.mydomain.com 
* email_server_port = 25
* email_username = sample_admin@mydomain.com
* email_password = abc
* email_from = admin <sample_admin@mydomain.com>  

**harbor_admin_password**: User account admin singned up by default this is the password of harbor admin, for example you can set Harbor12345  
**auth_mode**: By default the auth mode is db_auth, i.e. the credentials are stored in a databse. Please set it to ldap_auth if you want to verify user's credentials against an ldap server.  
**ldap_url**: The url for ldap endpoint, for example ldaps://ldap.mydomain.com  
**ldap_basedn**: The basedn template for verifying the user's password, for example uid=%s,ou=people,dc=mydomain,dc=com  
**db_password**: The password for root user of db, for example root123


## Manage Harbor
[Docker compose](https://docs.docker.com/compose/) is a tool for defining and running multi-container Docker applications  
`sudo docker-compose build` build or rebuild images according to docker-compose.yml  
`sudo docker-compose stop`  stop docker container according to docker-compose.yml  
`docker-compose rm -v`     remove docker container according to docker-compose.yml. Options "-v" remove volumes associated with containers  

