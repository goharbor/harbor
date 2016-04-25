# Installation and Configuration Guide of Harbor
Harbor can be installed by two approaches:  

1. Installing from the source code, which goes through a full build process. Internet connection is required.
2. Installing via a pre-built installation package, which saves time for building the code. Further, it provides a way to install Harbor to a host that is isolated from the Internet (offline installation).

This guide describes both approaches and their usage.

## Prerequisites of the target host
Harbor is deployed as several Docker containers. Hence, it can be deployed on any Linux distribution that supports Docker. 
Before deploying Harbor, the target host requires Python, Docker, Docker Compose to be installed.  
* Python should be version 2.7 or higher.  Some Linux distributions (Gentoo, Arch) may not have a Python interpreter installed by default. On those systems, you need to install Python manually.  
* The Docker engine should be version 1.10 or higher.  For the details to install Docker engine, please refer to: https://docs.docker.com/engine/installation/
* The Docker Compose needs to be version 1.6.0 or higher.  For the details to install Docker compose, please refer to: https://docs.docker.com/compose/install/

## Installing Harbor from the source code

To install from the source, the target host must be connected to the Internet.
#### Getting the source code:

```sh
$ git clone https://github.com/vmware/harbor
```
    
#### Configuring Harbor
Before installing Harbor, you should configure the parameters in the file **harbor.cfg**. You then execute the **prepare** script to generate configuration files for Harbor's containers. Finally, you use Docker Compose to start Harbor.  

At minimum, you need to change the **hostname** attribute in **harbor.cfg**. The description of each attribute is as follows:  

**hostname**: The hostname for a user to access the user interface and the registry service. It should be the IP address or the fully qualified domain name (FQDN) of your target machine, for example 192.168.1.10 or reg.yourdomain.com . Do NOT use localhost or 127.0.0.1 for the hostname because the registry service needs to be accessed by external clients.  
**ui_url_protocol**: The protocol for accessing the user interface and the token/notification service, by default it is http. To set up the https protocol, refer to [Configuring Harbor with HTTPS Access](configure_https.md).  
**Email settings**: the following 6 attributes are used to send an email to reset a user's password,  they are not mandatory unless the password reset function is needed in Harbor. By default SSL connection is not enabled, if your smtp server(such as exmail.qq.com) requires SSL connection and doesn't support STARTTLS, then you should enable it by set **email_ssl = true**.
* email_server = smtp.mydomain.com 
* email_server_port = 25
* email_username = sample_admin@mydomain.com
* email_password = abc
* email_from = admin <sample_admin@mydomain.com>  
* email_ssl = false

**harbor_admin_password**: The password for the administrator of Harbor, by default the password is Harbor12345, the user name is admin.  
**auth_mode**: The authentication mode of Harbor. By default it is *db_auth*, i.e. the credentials are stored in a database. Please set it to *ldap_auth* if you want to verify user's credentials against an LDAP server.  
**ldap_url**: The URL for LDAP endpoint, for example ldaps://ldap.mydomain.com. It is only used when **auth_mode** is set to *ldap_auth*.    
**ldap_basedn**: The basedn template for verifying the user's credentials against LDAP, for example uid=%s,ou=people,dc=mydomain,dc=com.  It is only used when **auth_mode** is set to *ldap_auth*.  
**db_password**: The password of root user of mySQL database. Change this password for any production use.  
**self_registration**: The flag to turn on or off the user self-registration function. If this flag is turned off, only an admin user can create new users in Harbor. The default value is on. 
NOTE: When **auth_mode** is *ldap_auth*, the self-registration feature is always disabled, therefore, this flag is ignored.  

#### Building and starting Harbor
After configuring harbor.cfg, build and start Harbor by the following commands. Because it requires downloading necessary files from the Internet, it may take a while for the docker-compose process to finish.  

```sh
    $ cd Deploy
    
    $ ./prepare
    Generated configuration file: ./config/ui/env
    Generated configuration file: ./config/ui/app.conf
    Generated configuration file: ./config/registry/config.yml
    Generated configuration file: ./config/db/env
    The configuration files are ready, please use docker-compose to start the service.

    $ sudo docker-compose up -d
```

If everything works fine, you can open a browser to visit the admin portal at http://reg.yourdomain.com . The default administrator username and password are admin/Harbor12345 .

Log in to the admin portal and create a new project, e.g. myproject. You can then use docker commands to login and push images. The default port of Harbor registry server is 80:
```sh
$ docker login reg.yourdomain.com
$ docker push reg.yourdomain.com/myproject/myrepo
```
**NOTE:** The default installation of Harbor uses HTTP protocol, you should add the option "--insecure-registry" to your client's Docker daemon and restart Docker service. 

For information on how to use Harbor, please refer to [User Guide of Harbor](user_guide.md) .

#### Configuring Harbor with HTTPS Access
Because Harbor does not ship with any certificates, it uses HTTP by default to serve registry requests. This makes it relatively simple to configure, especially for a development or testing environment. However, it is highly recommended that security be enabled for any production environment. Refer to [Configuring Harbor with HTTPS Access](configure_https.md) if you want to enable HTTPS access to Harbor.

## Installing Harbor via a pre-built installation package 

A pre-built installation package of each release can be downloaded from the [release page](https://github.com/vmware/harbor/releases). After downloading the package file **harbor-&lt;version&gt;.tgz** , extract files in the package.  
```
$ tar -xzvf harbor-0.1.1.tgz
$ cd harbor
```

Then configure Harbor by following instructions in Section [Configuring Harbor](#configuring-harbor). Next, run **prepare** script to generate config files and use docker compose to build Harbor's container images and eventually spin it up.


```
$ ./prepare
Generated configuration file: ./config/ui/env
Generated configuration file: ./config/ui/app.conf
Generated configuration file: ./config/registry/config.yml
Generated configuration file: ./config/db/env
The configuration files are ready, please use docker-compose to start the service.

$ sudo docker-compose up -d
......
```

### Deploying Harbor to a host which does not have Internet access
When you run *docker-compose up* to start Harbor, it will pull base images from Docker Hub and build new images for the containers. This process requires accessing the Internet. If you want to deploy Harbor to a host that is not connected to the Internet, you need to prepare Harbor on a machine that has access to the Internet. After that, you export the images as tgz files and transfer them to the target machine. Then load the tgz file into Docker's local image repo.

#### Building and saving images for offline installation
On a machine that is connected to the Internet, extract files from the pre-built installation package. Then run command "docker-compose build" to build the images and use the script *save_image.sh* to export them as tar files. The tar files will be stored in *images/* directory. Next, package everything in the directory *harbor/* into a tgz file and transfer it to the target machine. This can be done by executing the following commands:

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
$ tar -cvzf harbor_offline-0.1.1.tgz harbor
```

The file **harbor_offline-0.1.1.tgz** contains the images saved by previous steps and the other files required to start Harbor.
You can use tools such as scp to transfer the file **harbor_offline-0.1.1.tgz** to the target machine that does not have Internet connection. 
On the target machine, you can execute the following commands to start Harbor. Again, before running the **prepare** script, 
be sure to update **harbor.cfg** to reflect the right configuration of the target machine. (Refer to Section [Configuring Harbor](#configuring-harbor) .)
```
$ tar -xzvf harbor_offline-0.1.1.tgz  
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

### Managing Harbor's lifecycle
Harbor is composed of a few containers which are deployed via docker-compose, you can use docker-compose to manage the lifecycle of the containers. Below are a few useful commands: 

Build and start Harbor:  
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
Remove Harbor's containers while keeping the image data and Harbor's database files on the file system: 
```
$ sudo docker-compose rm
Going to remove harbor_proxy_1, harbor_ui_1, harbor_registry_1, harbor_mysql_1, harbor_log_1
Are you sure? [yN] y
Removing harbor_proxy_1 ... done
Removing harbor_ui_1 ... done
Removing harbor_registry_1 ... done
Removing harbor_mysql_1 ... done
```  

Remove Harbor's database and image data (for a clean re-installation):
```sh
$ rm -r /data/database
$ rm -r /data/registry
```

[Docker Compose command-line reference](https://docs.docker.com/compose/reference/) describes the usage information for the docker-compose subcommands.

### Persistent data and log files
By default, the data of database and image files in the registry are persisted in the directory **/data/** of the target machine. When Harbor's containers are removed and recreated, the data  remain unchanged. Harbor leverages rsyslog to collect the logs of each container, by default the log files are stored in the directory **/var/log/harbor/** on Harbor's host.  

##Troubleshooting
1.When setting up Harbor behind another nginx proxy or elastic load balancing, remove the below line if the proxy already has similar settings. Be sure to remove the line under these 3 sections: "location /", "location /v2/" and "location /service/".
```
proxy_set_header X-Forwarded-Proto $scheme;
```
