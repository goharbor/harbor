# Installation and Configuration Guide
Harbor can be installed by one of two approaches: 

- **Online installer:** The installer downloads Harbor's images from Docker hub. For this reason, the installer is very small in size.

- **Offline installer:** Use this installer when the host does not have an Internet connection. The installer contains pre-built images so its size is larger.


All installers can be downloaded from the **[official release](https://github.com/vmware/harbor/releases)** page. 

This guide describes the steps to install and configure Harbor by using the online or offline installer. The installation processes are almost the same. 

If you run a previous version of Harbor, you may need to migrate the data to fit the new database schema. For more details, please refer to **[Data Migration Guide](migration_guide.md)**.

In addition, the deployment instructions on Kubernetes has been created by the community. Refer to [Harbor on Kubernetes](kubernetes_deployment.md) for details.

## Prerequisites for the target host
Harbor is deployed as several Docker containers, and, therefore, can be deployed on any Linux distribution that supports Docker. The target host requires Python, Docker, and Docker Compose to be installed.  
* Python should be version 2.7 or higher.  Note that you may have to install Python on Linux distributions (Gentoo, Arch) that do not come with a Python interpreter installed by default  
* Docker engine should be version 1.10 or higher.  For installation instructions, please refer to: https://docs.docker.com/engine/installation/
* Docker Compose needs to be version 1.6.0 or higher.  For installation instructions, please refer to: https://docs.docker.com/compose/install/

## Installation Steps

The installation steps boil down to the following

1. Download the installer;
2. Configure **harbor.cfg**;
3. Run **install.sh** to install and start Harbor;


#### Downloading the installer:

The binary of the installer can be downloaded from the [release](https://github.com/vmware/harbor/releases) page. Choose either online or offline installer. Use *tar* command to extract the package.

Online installer:
```
    $ tar xvf harbor-online-installer-<version>.tgz
```
Offline installer:
```
    $ tar xvf harbor-offline-installer-<version>.tgz
```

#### Configuring Harbor
Configuration parameters are located in the file **harbor.cfg**. 

There are two categories of parameters in harbor.cfg, **required parameters** and **optional parameters**.  
* **required parameters**: These parameters are required to be set in the configuration file, and they will take effect if a user updates them in harbor.cfg, rerun the ```install.sh``` script to reinstall Harbor.
* **optional parameters**: These parameters are optional, and only take effect in the initial installation.  The user can leave them blank and update them on Web UI after Harbor is started. Subsequent update to these parameters in ```harbor.cfg``` will be ignored.

The parameters are described below - note that at the very least, you will need to change the **hostname** attribute. 

##### Required parameters:

* **hostname**: The target host's hostname, which is used to access the UI and the registry service. It should be the IP address or the fully qualified domain name (FQDN) of your target machine, e.g., `192.168.1.10` or `reg.yourdomain.com`. _Do NOT use `localhost` or `127.0.0.1` for the hostname - the registry service needs to be accessible by external clients!_ 
* **ui_url_protocol**: (**http** or **https**.  Default is **http**) The protocol used to access the UI and the token/notification service.  If Notary is enabled, this parameter has to be _https_.  By default, this is _http_. To set up the https protocol, refer to **[Configuring Harbor with HTTPS Access](configure_https.md)**.  
* **db_password**: The root password for the MySQL database used for **db_auth**. _Change this password for any production use!_ 
* **max_job_workers**: (default value is **3**) The maximum number of replication workers in job service. For each image replication job, a worker synchronizes all tags of a repository to the remote destination. Increasing this number allows more concurrent replication jobs in the system. However, since each worker consumes a certain amount of network/CPU/IO resources, please carefully pick the value of this attribute based on the hardware resource of the host. 
* **customize_crt**: (**on** or **off**.  Default is **on**) When this attribute is **on**, the prepare script creates private key and root certificate for the generation/verification of the registry's token. Set this attribute to **off** when the key and root certificate are supplied by external sources. Refer to [Customize Key and Certificate of Harbor Token Service](customize_token_service.md) for more info.
* **ssl_cert**: The path of SSL certificate, it's applied only when the protocol is set to https
* **ssl_cert_key**: The path of SSL key, it's applied only when the protocol is set to https 
* **secretkey_path**: The path of key for encrypt or decrypt the password of a remote registry in a replication policy.

##### Optional parameters
* **Email settings**: These parameters are needed for Harbor to be able to send a user a "password reset" email, and are only necessary if that functionality is needed.  Also, do note that by default SSL connectivity is _not_ enabled - if your SMTP server requires SSL, but does _not_ support STARTTLS, then you should enable SSL by setting **email_ssl = true**.
  * email_server = smtp.mydomain.com 
  * email_server_port = 25
  * email_username = sample_admin@mydomain.com
  * email_password = abc
  * email_from = admin <sample_admin@mydomain.com>  
  * email_ssl = false

* **harbor_admin_password**: The administrator's initial password. This password only takes effect for the first time Harbor launches. After that, this setting is ignored and the administrator's password should be set in the UI. _Note that the default username/password are **admin/Harbor12345** ._   
* **auth_mode**: The type of authentication that is used. By default, it is **db_auth**, i.e. the credentials are stored in a database. For LDAP authentication, set this to **ldap_auth**.  
* **ldap_url**: The LDAP endpoint URL (e.g. `ldaps://ldap.mydomain.com`).  _Only used when **auth_mode** is set to *ldap_auth* ._    
* **ldap_searchdn**: The DN of a user who has the permission to search an LDAP/AD server (e.g. `uid=admin,ou=people,dc=mydomain,dc=com`).
* **ldap_search_pwd**: The password of the user specified by *ldap_searchdn*.
* **ldap_basedn**: The base DN to look up a user, e.g. `ou=people,dc=mydomain,dc=com`.  _Only used when **auth_mode** is set to *ldap_auth* ._ 
* **ldap_filter**:The search filter for looking up a user, e.g. `(objectClass=person)`.
* **ldap_uid**: The attribute used to match a user during a LDAP search, it could be uid, cn, email or other attributes.
* **ldap_scope**: The scope to search for a user, 1-LDAP_SCOPE_BASE, 2-LDAP_SCOPE_ONELEVEL, 3-LDAP_SCOPE_SUBTREE. Default is 3. 
* **self_registration**: (**on** or **off**. Default is **on**) Enable / Disable the ability for a user to register himself/herself. When disabled, new users can only be created by the Admin user, only an admin user can create new users in Harbor.  _NOTE: When **auth_mode** is set to **ldap_auth**, self-registration feature is **always** disabled, and this flag is ignored._  
* **token_expiration**: The expiration time (in minutes) of a token created by token service, default is 30 minutes.
* **project_creation_restriction**: The flag to control what users have permission to create projects.  By default everyone can create a project, set to "adminonly" such that only admin can create project.
* **verify_remote_cert**: (**on** or **off**.  Default is **on**) This flag determines whether or not to verify SSL/TLS certificate when Harbor communicates with a remote registry instance. Setting this attribute to **off** bypasses the SSL/TLS verification, which is often used when the remote instance has a self-signed or untrusted certificate.

#### Configuring storage backend (optional)

By default, Harbor stores images on your local filesystem. In a production environment, you may consider 
using other storage backend instead of the local filesystem, like S3, Openstack Swift, Ceph, etc. 
What you need to update is the section of `storage` in the file `common/templates/registry/config.yml`. 
For example, if you use Openstack Swift as your storage backend, the section may look like this:

```
storage:
  swift:
    username: admin
    password: ADMIN_PASS
    authurl: http://keystone_addr:35357/v3/auth
    tenant: admin
    domain: default
    region: regionOne
    container: docker_images
```

_NOTE: For detailed information on storage backend of a registry, refer to [Registry Configuration Reference](https://docs.docker.com/registry/configuration/) ._


#### Finishing installation and starting Harbor
Once **harbor.cfg** and storage backend (optional) are configured, install and start Harbor using the ```install.sh``` script.  Note that it may take some time for the online installer to download Harbor images from Docker hub.  

##### Default installation
After version 1.1.0, Harbor has integrated with Notary, but by default the installation does not include notary support.

```sh
    $ sudo ./install.sh
```

If everything worked properly, you should be able to open a browser to visit the admin portal at **http://reg.yourdomain.com** (change *reg.yourdomain.com* to the hostname configured in your harbor.cfg). Note that the default administrator username/password are admin/Harbor12345 .

Log in to the admin portal and create a new project, e.g. `myproject`. You can then use docker commands to login and push images (By default, the registry server listens on port 80):
```sh
$ docker login reg.yourdomain.com
$ docker push reg.yourdomain.com/myproject/myrepo:mytag
```
**IMPORTANT:** The default installation of Harbor uses _HTTP_ - as such, you will need to add the option `--insecure-registry` to your client's Docker daemon and restart the Docker service. 

##### Installation with Notary
To install Harbor with Notary support, add a parameter when you run ```install.sh```
```sh
    $ sudo ./install.sh --with-notary
```
**Note**: For installation with Notary the parameter "ui_url_protocol" must be set to "https", for configuring HTTPS certificate please refer to the following sections.

More information about Notary and Docker Content Trust, please refer to docker's documentation: 
https://docs.docker.com/engine/security/trust/content_trust/

For information on how to use Harbor, please refer to **[User Guide of Harbor](user_guide.md)** .

#### Configuring Harbor with HTTPS access
Harbor does not ship with any certificates, and, by default, uses HTTP to serve requests. While this makes it relatively simple to set up and run - especially for a development or testing environment - it is **not** recommended for a production environment.  To enable HTTPS, please refer to **[Configuring Harbor with HTTPS Access](configure_https.md)**.  


### Managing Harbor's lifecycle
You can use docker-compose to manage the lifecycle of Harbor. Some useful commands are listed as follows (must run in the same directory as *docker-compose.yml*).

Stopping Harbor:
```
$ sudo docker-compose stop
Stopping nginx ... done
Stopping harbor-jobservice ... done
Stopping harbor-ui ... done
Stopping harbor-db ... done
Stopping registry ... done
Stopping harbor-log ... done
```  
Restarting Harbor after stopping:
```
$ sudo docker-compose start
Starting log ... done
Starting ui ... done
Starting mysql ... done
Starting jobservice ... done
Starting registry ... done
Starting proxy ... done
```  

To change Harbor's configuration, first stop existing Harbor instance, update harbor.cfg, and then run prepare script to populate the configuration, and then re-create and start Harbor's instance:
```
$ sudo docker-compose down -v
$ vim harbor.cfg
$ sudo prepare
$ sudo docker-compose up -d
``` 

Removing Harbor's containers while keeping the image data and Harbor's database files on the file system:
```
$ sudo docker-compose down -v
```  

Removing Harbor's database and image data (for a clean re-installation):
```sh
$ rm -r /data/database
$ rm -r /data/registry
```

#### _Managing lifecycle of Harbor when it's installed with Notary_ 

When Harbor is installed with Notary, user needs to add extra template file ```docker-compose.notary.yml``` to docker-compose command, so the docker-compose commands to manage the lifecycle of Harbor will be:
```
$ sudo docker-compose -f ./docker-compose.yml -f ./docker-compose.notary.yml [ up|down|ps|stop|start ]
```
For example, if user want's to change ```harbor.cfg``` and re-deploy Harbor when it's installed with Notary, the following commands should be used:
```sh
$ sudo docker-compose -f ./docker-compose.yml -f ./docker-compose.notary.yml down -v
$ vim harbor.cfg
$ sudo prepare --with-notary
$ sudo docker-compose -f ./docker-compose.yml -f ./docker-compose.notary.yml up -d
```

Please check the [Docker Compose command-line reference](https://docs.docker.com/compose/reference/) for more on docker-compose.

### Persistent data and log files
By default, registry data is persisted in the target host's `/data/` directory.  This data remains unchanged even when Harbor's containers are removed and/or recreated.  

In addition, Harbor uses *rsyslog* to collect the logs of each container. By default, these log files are stored in the directory `/var/log/harbor/` on the target host for troubleshooting.  

## Configuring Harbor listening on a customized port
By default, Harbor listens on port 80(HTTP) and 443(HTTPS, if configured) for both admin portal and docker commands, you can configure it with a customized one.  

### For HTTP protocol

1.Modify docker-compose.yml  
Replace the first "80" to a customized port, e.g. 8888:80.  

```
proxy:
    image: library/nginx:1.11.5
    restart: always
    volumes:
      - ./config/nginx:/etc/nginx
    ports:
      - 8888:80
      - 443:443
    depends_on:
      - mysql
      - registry
      - ui
      - log
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "proxy"
```

2.Modify harbor.cfg, add the port to the parameter "hostname"  

```  
hostname = 192.168.0.2:8888
```

3.Run install.sh to update and start Harbor.  
```sh
$ sudo docker-compose down
$ sudo install.sh
```
### For HTTPS protocol
1.Enable HTTPS in Harbor by following this [guide](https://github.com/vmware/harbor/blob/master/docs/configure_https.md).  
2.Modify docker-compose.yml  
Replace the first "443" to a customized port, e.g. 4443:443.  

```
proxy:
    image: library/nginx:1.11.5
    restart: always
    volumes:
      - ./config/nginx:/etc/nginx
    ports:
      - 80:80
      - 4443:443
    depends_on:
      - mysql
      - registry
      - ui
      - log
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "proxy"
```

3.Modify harbor.cfg, add the port to the parameter "hostname"  

```  
hostname = 192.168.0.2:8888
```

4.Run install.sh to update and start Harbor.  
```sh
$ sudo docker-compose down -v
$ sudo prepare
$ sudo docker-compose up -d
```


## Troubleshooting
1. When Harbor does not work properly, run the below commands to find out if all containers of Harbor are in **UP** status: 
```
    $ sudo docker-compose ps
        Name                     Command               State                    Ports                   
  -----------------------------------------------------------------------------------------------------
  harbor-db           docker-entrypoint.sh mysqld      Up      3306/tcp                                 
  harbor-jobservice   /harbor/harbor_jobservice        Up                                               
  harbor-log          /bin/sh -c crond && rsyslo ...   Up      127.0.0.1:1514->514/tcp                    
  harbor-ui           /harbor/harbor_ui                Up                                               
  nginx               nginx -g daemon off;             Up      0.0.0.0:443->443/tcp, 0.0.0.0:80->80/tcp 
  registry            /entrypoint.sh serve /etc/ ...   Up      5000/tcp                                 
```
If a container is not in **UP** state, check the log file of that container in directory ```/var/log/harbor```. For example, if the container ```harbor-ui``` is not running, you should look at the log file ```ui.log```.  


2.When setting up Harbor behind an nginx proxy or elastic load balancing, look for the line below, in `common/templates/nginx/nginx.http.conf` and remove it from the sections if the proxy already has similar settings: `location /`, `location /v2/` and `location /service/`.
```
proxy_set_header X-Forwarded-Proto $scheme;
```
And run the following commands to restart Harbor:
```sh
$ sudo docker-compose down -v
$ sudo ./prepare
$ sudo docker-compose up -d
```
