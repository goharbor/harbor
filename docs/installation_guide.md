# Harbor Installation 
### Download the installation package
The installation package file for each release can be downloaded from the [release tab](https://github.com/vmware/harbor/releases).
### Prerequisites for target machine
Harbor is deployed as several docker containers.  Hence, it can be deployed on any Linux distribution. 
Before deploying harbor, the target machine requires python, docker-engine, docker compose installede.  
* Python needs to be verion 2.7 or higher.  Some Linux distributions (Gentoo, Arch) may not have a Python interpreter installed by default. On those systems, you need to install one.  
* The docker engine needs to be version 1.8 or higher.  For the details to install docker engine, please refer to: https://docs.docker.com/engine/installation/
* The docker-compose needs to be version 1.6.0 or higher.  For the details to install docker compose, please refer to: https://docs.docker.com/compose/install/

### Configure Harbor Parameters 
Taking version 0.1.0 as an example, after downloading the package file **harbor-0.1.0.tgz** from release tab, you need to extract the package, make update to the parameters in the file **harbor.cfg**, execute the **prepare** script to generate configuration files for docker containers, and use docker-compose to start the service.  
For simplest case, you only need to update the **hostname** attribute in **harbor.cfg** by updating the value to the IP or fully qualified hostname of your target machine, for example 192.168.1.10.  Please see the next section for the details of each attriubte.
```
$ tar -xzvf harbor-0.1.0.tgz
$ cd harbor
$ vi ./harbor.cfg
......
$ ./prepare
Generated configuration file: ./config/ui/env
Generated configuration file: ./config/ui/app.conf
Generated configuration file: ./config/registry/config.yml
Generated configuration file: ./config/db/env
The configuration files are ready, please use docker-compose to start the service.
$ sudo docker-compose up -d
......
```
After that, you can open your browser and access harbor via the IP you set in harbor.cfg, such as 192.168.1.10
#### Parameters in harbor.cfg
**hostname**: The endpoint for user to access UI and registry service, for example 192.168.1.10 or exampledomian.com.  
**ui_url_protocol**: The protocol for accessing the UI and token/notification service, by default it is http.  
**Email settings**: the following 5 attributes are used to send password resetting email, by default it is not necessary.  
* email_server = smtp.mydomain.com 
* email_server_port = 25
* email_username = sample_admin@mydomain.com
* email_password = abc
* email_from = admin <sample_admin@mydomain.com>  

**harbor_admin_password**: The password for administrator of harbor, by default it is Harbor12345.  
**auth_mode**: The authentication mode of harbor.  By default the it is *db_auth*, i.e. the credentials are stored in a databse. Please set it to *ldap_auth* if you want to verify user's credentials against an LDAP server.  
**ldap_url**: The URL for LDAP endpoint, for example ldaps://ldap.mydomain.com. It is only used when **auth_mode** is set as *ldap_auth*.    
**ldap_basedn**: The basedn template for verifying the user's credentials against LDAP, for example uid=%s,ou=people,dc=mydomain,dc=com.  It is only used when **auth_mode** set as *ldap_auth*.  
**db_password**: The password of root user of mySQL database.

### Deploy harbor to a target machine that does not have internet access
When you run *docker-compose up* to start harbor service.  Docker will pull base images from docker hub and build new images for the containers.  This process requires accessing internet.  If you want to deploy harbor to a target machine in intranet which does not have access to the internet, essentially you need to first export the images as tgz files and transfer them to the target machine, then load the tgz file as docker images.

#### Build and save service images
After extracting the installation package.  Use command "docker-compose build" to build the images and run the script *save_image.sh* to export them as tar files and they will be stored in **images** directory, after that, user can package everything in directory **harbor** into a tgz file and transfer the tgz file to target machine.  This can be done by executing the following commands:

```
$ cd harbor
$ sudo docker-compose build
......
$ sudo ./save_image.sh  
saving the image of harbor_ui
finished saving the image of harbor_ui
saving the image of harbor_log
finished saving the image of harbor_log
saving the image of harbor_mysql
finished saving the image of harbor_mysql
saving the image of nginx
finished saving the image of nginx
saving the image of registry
finished saving the image of registry
$ cd ../  
$ tar -cvzf harbor_offline-0.1.0.tgz harbor
```

The package file **harbor_offline-0.1.0.tgz** contains the images saved by previously steps and the files needed to start harbor services.
Then you can use tools such as scp to transfer the file **harbor_offline-0.1.0.tgz** to the target machine that does not have internet access.  Then on the target machine, you can execute the following commands to start harbor service.
```
$ tar -xzvf harbor_offline-0.1.tgz  
$ cd harbor  
# load images save by excute ./save_image.sh
$ ./load_image.sh
loading the image of harbor_ui
finish loaded the image of harbor_ui
loading the image of harbor_mysql
finished loading the image of harbor_mysql
loading the image of nginx
finished loading the image of nginx
loading the image of registry
finished loading the image of registry
# Make update to the parameters in ./harbor.cfg  
$ ./prepare
Generated configuration file: ./config/ui/env
Generated configuration file: ./config/ui/app.conf
Generated configuration file: ./config/registry/config.yml
Generated configuration file: ./config/db/env
The configuration files are ready, please use docker-compose to start the service.
# Build the images and then start the services
$ sudo docker-compose up -d
```

### Manage Harbor Lifecycle
Harbor are deployed via docker-compose, you can use docker-compose to manage the lifecycle of the containers as a group.  Below are a few useful commands:  
create and start containers according to docker-compose.yml  
```
$ sudo docker-compose up -d 
Creating harbor_log_1
Creating harbor_mysql_1
Creating harbor_registry_1
Creating harbor_ui_1
Creating harbor_proxy_1
```  
stop docker container according to docker-compose.yml  
```
$ sudo docker-compose stop
Stopping harbor_proxy_1 ... done
Stopping harbor_ui_1 ... done
Stopping harbor_registry_1 ... done
Stopping harbor_mysql_1 ... done
Stopping harbor_log_1 ... done
```  
start stopped services according to docker-compose.yml  
```
$ sudo docker-compose start
Starting harbor_log_1
Starting harbor_mysql_1
Starting harbor_registry_1
Starting harbor_ui_1
Starting harbor_proxy_1
````  
remove stopped containers  
```
$ sudo docker-compose rm
Going to remove harbor_proxy_1, harbor_ui_1, harbor_registry_1, harbor_mysql_1, harbor_log_1
Are you sure? [yN] y
Removing harbor_proxy_1 ... done
Removing harbor_ui_1 ... done
Removing harbor_registry_1 ... done
Removing harbor_mysql_1 ... done
```  
[Compose command-line reference](https://docs.docker.com/compose/reference/) describes the usage information for the docker-compose subcommands.

### Persistent data and log files
By default, data of database, and image files in registry are persisted in directory **/data/** of the target machine.  When the containers are removed and recreated the data will remain unchanged.  
Harbor leverage rsyslog to collect the logs of each container, by default the log files are stored in directory **/var/log/harbor/**