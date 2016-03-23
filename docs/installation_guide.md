# Installation Guide of Harbor
### Download the installation package
Harbor can be installed from the source code by using "docker-compose up" command, which goes through a full build process. Besides, a pre-built installation package for each release can be downloaded from the [release page](https://github.com/vmware/harbor/releases). This guide describes the installation of Harbor by the pre-built package.
### Prerequisites for target machine
Harbor is deployed as several Docker containers.  Hence, it can be deployed on any Linux distribution that supports Docker. 
Before deploying Harbor, the target machine requires Python, Docker, Docker Compose to be installed.  
* Python should be version 2.7 or higher.  Some Linux distributions (Gentoo, Arch) may not have a Python interpreter installed by default. On those systems, you need to install Python manually.  
* The Docker engine should be version 1.8 or higher.  For the details to install Docker engine, please refer to: https://docs.docker.com/engine/installation/
* The Docker Compose needs to be version 1.6.0 or higher.  For the details to install Docker compose, please refer to: https://docs.docker.com/compose/install/

### Configuration of Harbor 
After downloading the package file **```harbor-<version>.tgz```** from release page, you need to extract the package. Before installing Harbor, configure the parameters in the file **harbor.cfg**. Then execute the **prepare** script to generate configuration files for Harbor's containers. Finally, use Docker Compose to start the service.  
At minimum, you only need to change the **hostname** attribute in **harbor.cfg** by updating the IP  address or fully qualified hostname of your target machine, for example 192.168.1.10.  Please see the next section for the description of each parameter.
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
After that, you can open a browser and access Harbor via the IP you set in harbor.cfg, such as http://192.168.1.10 . The same IP address is used as the Registry address in your Docker client, for example:  
```docker pull 192.168.1.10/library/ubuntu```

#### Parameters in harbor.cfg
**hostname**: The endpoint for user to access UI and registry service, for example 192.168.1.10 or exampledomian.com.  
**ui_url_protocol**: The protocol for accessing the UI and token/notification service, by default it is http.  
**Email settings**: the following 5 attributes are used to send an email to reset user's password,  it is not mandatory unless password reset function is needed in Harbor.  
* email_server = smtp.mydomain.com 
* email_server_port = 25
* email_username = sample_admin@mydomain.com
* email_password = abc
* email_from = admin <sample_admin@mydomain.com>  

**harbor_admin_password**: The password for administrator of Harbor, by default it is Harbor12345, the user name is admin.  
**auth_mode**: The authentication mode of Harbor. By default it is *db_auth*, i.e. the credentials are stored in a database. Please set it to *ldap_auth* if you want to verify user's credentials against an LDAP server.  
**ldap_url**: The URL for LDAP endpoint, for example ldaps://ldap.mydomain.com. It is only used when **auth_mode** is set to *ldap_auth*.    
**ldap_basedn**: The basedn template for verifying the user's credentials against LDAP, for example uid=%s,ou=people,dc=mydomain,dc=com.  It is only used when **auth_mode** is set to *ldap_auth*.  
**db_password**: The password of root user of mySQL database.

### Deploy Harbor to a target machine that does not have Internet access
When you run *docker-compose up* to start Harbor service. It will pull base images from Docker hub and build new images for the containers. This process requires accessing the Internet. If you want to deploy Harbor to a host that is not connected to the Internet, you need to prepare Harbor on a machine that has access to the Internet. After that, you export the images as tgz files and transfer them to the target machine, then load the tgz file into Docker's local image repo.

#### Build and save images for offline installation
On a machine that is connect to Internet, extract the installation package. Then run command "docker-compose build" to build the images and use the script *save_image.sh* to export them as tar files. The tar files will be stored in **images** directory. Next, user can package everything in directory **harbor** into a tgz file and transfer the tgz file to the target machine. This can be done by executing the following commands:

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

The package file **harbor_offline-0.1.0.tgz** contains the images saved by previously steps and the files needed to start Harbor services.
Then you can use tools such as scp to transfer the file **harbor_offline-0.1.0.tgz** to the target machine that does not have Internet access. On the target machine, you can execute the following commands to start Harbor service. Again, before running the **prepare** script, be sure to update **harbor.cfg** to reflect the right configuration of the target machine.
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

### Manage Harbor's lifecycle
Harbor is composed of a few containers which are deployed via docker-compose, you can use docker-compose to manage the lifecycle of the containers. Below are a few useful commands: 

Create and start Harbor:  
```
$ sudo docker-compose up -d 
Creating harbor_log_1
Creating harbor_mysql_1
Creating harbor_registry_1
Creating harbor_ui_1
Creating harbor_proxy_1
```  
Stop Harbor:
```
$ sudo docker-compose stop
Stopping harbor_proxy_1 ... done
Stopping harbor_ui_1 ... done
Stopping harbor_registry_1 ... done
Stopping harbor_mysql_1 ... done
Stopping harbor_log_1 ... done
```  
Restart Harbor after stopping
```
$ sudo docker-compose start
Starting harbor_log_1
Starting harbor_mysql_1
Starting harbor_registry_1
Starting harbor_ui_1
Starting harbor_proxy_1
````  
Remove Harbor's containers (the image data and Harbor database files remains on the file system): 
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
By default, data of database and image files in registry are persisted in directory **/data/** of the target machine. When Harbor's containers are removed and recreated the data will remain unchanged. 
Harbor leverages rsyslog to collect the logs of each container, by default the log files are stored in directory **/var/log/harbor/** .