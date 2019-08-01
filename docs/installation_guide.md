# Installation and Configuration Guide

Harbor can be installed by one of two approaches:

- **Online installer:** The installer downloads Harbor's images from Docker hub. For this reason, the installer is very small in size.

- **Offline installer:** Use this installer when the host does not have an Internet connection. The installer contains pre-built images so its size is larger.

All installers can be downloaded from the **[official release](https://github.com/goharbor/harbor/releases)** page.

This guide describes the steps to install and configure Harbor by using the online or offline installer. The installation processes are almost the same.

If you run a previous version of Harbor, you may need to update ```harbor.yml``` and migrate the data to fit the new database schema. For more details, please refer to **[Harbor Migration Guide](migration_guide.md)**.

In addition, the deployment instructions on Kubernetes has been created by the community. Refer to [Harbor on Kubernetes](kubernetes_deployment.md) for details.

## Prerequisites for the target host

Harbor is deployed as several Docker containers, and, therefore, can be deployed on any Linux distribution that supports Docker. The target host requires Docker, and Docker Compose to be installed.

### Hardware

|Resource|Capacity|Description|
|---|---|---|
|CPU|minimal 2 CPU|4 CPU is preferred|
|Mem|minimal 4GB|8GB is preferred|
|Disk|minimal 40GB|160GB is preferred|

### Software

|Software|Version|Description|
|---|---|---|
|Docker engine|version 17.06.0-ce+ or higher|For installation instructions, please refer to: [docker engine doc](https://docs.docker.com/engine/installation/)|
|Docker Compose|version 1.18.0 or higher|For installation instructions, please refer to: [docker compose doc](https://docs.docker.com/compose/install/)|
|Openssl|latest is preferred|Generate certificate and keys for Harbor|

### Network ports

|Port|Protocol|Description|
|---|---|---|
|443|HTTPS|Harbor portal and core API will accept requests on this port for https protocol, this port can change in config file|
|4443|HTTPS|Connections to the Docker Content Trust service for Harbor, only needed when Notary is enabled, This port can change in config file|
|80|HTTP|Harbor portal and core API will accept requests on this port for http protocol|

## Installation Steps

The installation steps boil down to the following

1. Download the installer;
2. Configure **harbor.yml**;
3. Run **install.sh** to install and start Harbor;

#### Downloading the installer:

The binary of the installer can be downloaded from the [release](https://github.com/goharbor/harbor/releases) page. Choose either online or offline installer. Use *tar* command to extract the package.

Online installer:

```bash
    $ tar xvf harbor-online-installer-<version>.tgz
```

Offline installer:

```bash
    $ tar xvf harbor-offline-installer-<version>.tgz
```

#### Configuring Harbor

Configuration parameters are located in the file **harbor.yml**.

There are two categories of parameters, **required parameters** and **optional parameters**.

- **System level parameters**: These parameters are required to be set in the configuration file. They will take effect if a user updates them in ```harbor.yml``` and run the ```install.sh``` script to reinstall Harbor.

- **User level parameters**: These parameters can update after the first time harbor started on Web Portal. In particular, you must set the desired **auth_mode** before registering or creating any new users in Harbor. When there are users in the system (besides the default admin user), **auth_mode** cannot be changed.

The parameters are described below - note that at the very least, you will need to change the **hostname** attribute.

##### Required parameters

- **hostname**: The target host's hostname, which is used to access the Portal and the registry service. It should be the IP address or the fully qualified domain name (FQDN) of your target machine, e.g., `192.168.1.10` or `reg.yourdomain.com`. _Do NOT use `localhost` or `127.0.0.1` for the hostname - the registry service needs to be accessible by external clients!_

- **data_volume**: The location to store harbor's data.

- **harbor_admin_password**: The administrator's initial password. This password only takes effect for the first time Harbor launches. After that, this setting is ignored and the administrator's password should be set in the Portal. _Note that the default username/password are **admin/Harbor12345** ._



- **database**: the configs related to local database
  - **password**: The root password for the PostgreSQL database used for **db_auth**. _Change this password for any production use!_

- **jobservice**: jobservice related service
  - **max_job_workers**: The maximum number of replication workers in job service. For each image replication job, a worker synchronizes all tags of a repository to the remote destination. Increasing this number allows more concurrent replication jobs in the system. However, since each worker consumes a certain amount of network/CPU/IO resources, please carefully pick the value of this attribute based on the hardware resource of the host.
- **log**: log related url
  - **level**: log level, options are debug, info, warning, error, fatal
  - **rotate_count**: Log files are rotated **rotate_count** times before being removed. If count is 0, old versions are removed rather than rotated.
  - **rotate_size**: Log files are rotated only if they grow bigger than **rotate_size** bytes. If size is followed by k, the size is assumed to be in kilobytes. If the M is used, the size is in megabytes, and if G is used, the size is in gigabytes. So size 100, size 100k, size 100M and size 100G are all valid.
  - **location**: he directory to store store log

##### optional parameters

- **http**:
  - **port** : the port number of you http

- **https**: The protocol used to access the Portal and the token/notification service.  If Notary is enabled, has to set to _https_.
refer to **[Configuring Harbor with HTTPS Access](configure_https.md)**.
  - **port**: port number for https
  - **certificate**: The path of SSL certificate, it's applied only when the protocol is set to https.
  - **private_key**: The path of SSL key, it's applied only when the protocol is set to https.

- **external_url**: Enable it if use external proxy, and when it enabled the hostname will no longer used

- **clair**: Clair related configs
  - **updaters_interval**: The interval of clair updaters, the unit is hour, set to 0 to disable the updaters
  - **http_proxy**: Config http proxy for Clair, e.g. `http://my.proxy.com:3128`.
  - **https_proxy**: Config https proxy for Clair, e.g. `http://my.proxy.com:3128`.
  - **no_proxy**: Config no proxy for Clair, e.g. `127.0.0.1,localhost,core,registry`.

- **chart**: chart related configs
  - **absolute_url**: if set to enabled chart will use absolute url, otherwise set it to disabled, chart will use relative url.

- **external_database**: external database configs, Currently only support POSTGRES.
  - **harbor**: harbor's core database configs
    - **host**: hostname for harbor core database
    - **port**: port of harbor's core database
    - **db_name**: database name of harbor core database
    - **username**: username to connect harbor core database
    - **password**: password to harbor core database
    - **ssl_mode**: is enable ssl mode
  - **clair**: clair's database configs
    - **host**: hostname for clair database
    - **port**: port of clair database
    - **db_name**: database name of clair database
    - **username**: username to connect clair database
    - **password**: password to clair database
    - **ssl_mode**: is enable ssl mode
  - **notary_signer**: notary's signer database configs
    - **host**: hostname for notary signer database
    - **port**: port of notary signer database
    - **db_name**: database name of notary signer database
    - **username**: username to connect notary signer database
    - **password**: password to notary signer database
    - **ssl_mode**: is enable ssl mode
  - **notary_server**:
    - **host**: hostname for notary server database
    - **port**: port of notary server database
    - **db_name**: database name of notary server database
    - **username**: username to connect notary server database
    - **password**: password to notary server database
    - **ssl_mode**: is enable ssl mode

- **external_redis**: configs for use the external redis
  - **host**: host for external redis
  - **port**: port for external redis
  - **password**: password to connect external host
  - **registry_db_index**: db index for registry use
  - **jobservice_db_index**: db index for jobservice
  - **chartmuseum_db_index**: db index for chartmuseum

#### Configuring storage backend (optional)

- **storage_service**: By default, Harbor stores images and chart on your local filesystem. In a production environment, you may consider use other storage backend instead of the local filesystem, like S3, OpenStack Swift, Ceph, etc. These parameters are configurations for registry.
  - **ca_bundle**:  The path to the custom root ca certificate, which will be injected into the trust store of registry's and chart repository's containers.  This is usually needed when the user hosts a internal storage with self signed certificate.
  - **provider_name**: Storage configs for registry, default is filesystem. for more info about this configuration please refer https://docs.docker.com/registry/configuration/
  - **redirect**:
    - **disable**: set disable to true when you want to disable registry redirect

For example, if you use Openstack Swift as your storage backend, the parameters may look like this:

``` yaml
storage_service:
  ca_bundle:
  swift:
    username: admin
    password: ADMIN_PASS
    authurl: http://keystone_addr:35357/v3/auth
    tenant: admin
    domain: default
    region: regionOne
    container: docker_images"
  redirect:
    disable: false
```

_NOTE: For detailed information on storage backend of a registry, refer to [Registry Configuration Reference](https://docs.docker.com/registry/configuration/) ._

#### Finishing installation and starting Harbor

Once **harbor.yml** and storage backend (optional) are configured, install and start Harbor using the `install.sh` script.  Note that it may take some time for the online installer to download Harbor images from Docker hub.

##### Default installation (without Notary/Clair)

Harbor has integrated with Notary and Clair (for vulnerability scanning). However, the default installation does not include Notary or Clair service.

``` sh
    $ sudo ./install.sh
```

If everything worked properly, you should be able to open a browser to visit the admin portal at `http://reg.yourdomain.com` (change `reg.yourdomain.com` to the hostname configured in your `harbor.yml`). Note that the default administrator username/password are admin/Harbor12345.

Log in to the admin portal and create a new project, e.g. `myproject`. You can then use docker commands to login and push images (By default, the registry server listens on port 80):

```sh
$ docker login reg.yourdomain.com
$ docker push reg.yourdomain.com/myproject/myrepo:mytag
```

**IMPORTANT:** The default installation of Harbor uses _HTTP_ - as such, you will need to add the option `--insecure-registry` to your client's Docker daemon and restart the Docker service.

##### Installation with Notary
To install Harbor with Notary service, add a parameter when you run `install.sh`:

```sh
    $ sudo ./install.sh --with-notary
```

**Note**: For installation with Notary the parameter **ui_url_protocol** must be set to "https". For configuring HTTPS please refer to the following sections.

More information about Notary and Docker Content Trust, please refer to [Docker's documentation](https://docs.docker.com/engine/security/trust/content_trust/).

##### Installation with Clair

To install Harbor with Clair service, add a parameter when you run `install.sh`:

```sh
    $ sudo ./install.sh --with-clair
```

For more information about Clair, please refer to Clair's documentation:
`https://coreos.com/clair/docs/2.0.1/`

##### Installation with chart repository service

To install Harbor with chart repository service, add a parameter when you run ```install.sh```:

```sh
    $ sudo ./install.sh --with-chartmuseum
```

**Note**: If you want to install Notary, Clair and chart repository service, you must specify all the parameters in the same command:

```sh
    $ sudo ./install.sh --with-notary --with-clair --with-chartmuseum
```

For information on how to use Harbor, please refer to **[User Guide of Harbor](user_guide.md)** .

#### Configuring Harbor with HTTPS access

Harbor does not ship with any certificates, and, by default, uses HTTP to serve requests. While this makes it relatively simple to set up and run - especially for a development or testing environment - it is **not** recommended for a production environment.  To enable HTTPS, please refer to **[Configuring Harbor with HTTPS Access](configure_https.md)**.

### Managing Harbor's lifecycle

You can use docker-compose to manage the lifecycle of Harbor. Some useful commands are listed as follows (must run in the same directory as *docker-compose.yml*).

Stopping Harbor:

``` sh
$ sudo docker-compose stop
Stopping nginx              ... done
Stopping harbor-portal      ... done
Stopping harbor-jobservice  ... done
Stopping harbor-core        ... done
Stopping registry           ... done
Stopping redis              ... done
Stopping registryctl        ... done
Stopping harbor-db          ... done
Stopping harbor-log         ... done
```

Restarting Harbor after stopping:

``` sh
$ sudo docker-compose start
Starting log         ... done
Starting registry    ... done
Starting registryctl ... done
Starting postgresql  ... done
Starting core        ... done
Starting portal      ... done
Starting redis       ... done
Starting jobservice  ... done
Starting proxy       ... done
```

To change Harbor's configuration, first stop existing Harbor instance and update `harbor.yml`. Then run `prepare` script to populate the configuration. Finally re-create and start Harbor's instance:

``` sh
$ sudo docker-compose down -v
$ vim harbor.yml
$ sudo prepare
$ sudo docker-compose up -d
```

Removing Harbor's containers while keeping the image data and Harbor's database files on the file system:

``` sh
$ sudo docker-compose down -v
```

Removing Harbor's database and image data (for a clean re-installation):

``` sh
$ rm -r /data/database
$ rm -r /data/registry
```

#### *Managing lifecycle of Harbor when it's installed with Notary, Clair and chart repository service*

If you want to install Notary, Clair and chart repository service together, you should include all the components in the prepare commands:

``` sh
$ sudo docker-compose down -v
$ vim harbor.yml
$ sudo prepare --with-notary --with-clair --with-chartmuseum
$ sudo docker-compose up -d
```

Please check the [Docker Compose command-line reference](https://docs.docker.com/compose/reference/) for more on docker-compose.

### Persistent data and log files

By default, registry data is persisted in the host's `/data/` directory.  This data remains unchanged even when Harbor's containers are removed and/or recreated, you can edit the `data_volume` in `harbor.yml` file to change this directory.

In addition, Harbor uses *rsyslog* to collect the logs of each container. By default, these log files are stored in the directory `/var/log/harbor/` on the target host for troubleshooting, also you can change the log directory in `harbor.yml`.

## Configuring Harbor listening on a customized port

By default, Harbor listens on port 80(HTTP) and 443(HTTPS, if configured) for both admin portal and docker commands, these default ports can configured in `harbor.yml`

## Configuring Harbor using the external database

Currently, only PostgreSQL database is supported by Harbor.
To user an external database, just uncomment the `external_database` section in `harbor.yml` and fill the necessary information. Four databases are needed to be create first by users for Harbor core, Clair, Notary server and Notary signer. And the tables will be generated automatically when Harbor starting up.

## Manage user settings

After release 1.8.0, User settings are separated with system settings, and all user settings should be configured in web console or by HTTP request.
Please refer [Configure User Settings](configure_user_settings.md) to config user settings.

## Performance tuning

By default, Harbor limits the CPU usage of Clair container to 150000 and avoids its using up all the CPU resources. This is defined in the docker-compose.clair.yml file. You can modify it based on your hardware configuration.

## Troubleshooting

1. When Harbor does not work properly, run the below commands to find out if all containers of Harbor are in **UP** status:
```
    $ sudo docker-compose ps
        Name                     Command               State                    Ports
  -----------------------------------------------------------------------------------------------------------------------------
  harbor-core         /harbor/start.sh                 Up
  harbor-db           /entrypoint.sh postgres          Up      5432/tcp
  harbor-jobservice   /harbor/start.sh                 Up
  harbor-log          /bin/sh -c /usr/local/bin/ ...   Up      127.0.0.1:1514->10514/tcp
  harbor-portal       nginx -g daemon off;             Up      80/tcp
  nginx               nginx -g daemon off;             Up      0.0.0.0:443->443/tcp, 0.0.0.0:4443->4443/tcp, 0.0.0.0:80->80/tcp
  redis               docker-entrypoint.sh redis ...   Up      6379/tcp
  registry            /entrypoint.sh /etc/regist ...   Up      5000/tcp
  registryctl         /harbor/start.sh                 Up
```

If a container is not in **UP** state, check the log file of that container in directory `/var/log/harbor`. For example, if the container `harbor-core` is not running, you should look at the log file `core.log`.

2.When setting up Harbor behind an nginx proxy or elastic load balancing, look for the line below, in `common/config/nginx/nginx.conf` and remove it from the sections if the proxy already has similar settings: `location /`, `location /v2/` and `location /service/`.

``` sh
proxy_set_header X-Forwarded-Proto $scheme;
```

and re-deploy Harbor refer to the previous section "Managing Harbor's lifecycle".
