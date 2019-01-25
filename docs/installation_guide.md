# Installation and Configuration Guide

Harbor can be installed using one of two available approaches: 

- **Online installer:** This installer downloads Harbor's Docker images from [Docker Hub](https://hub.docker.com) and is thus very small in size. Use this installer when the host has access to the Internet.

- **Offline installer:** This installer contains pre-built images and is thus much larger than the online installer. Use this installer when the host doesn't have an Internet connection.

Both installers can be downloaded from Harbor's **[official releases](https://github.com/goharbor/harbor/releases)** page.

This guide describes the steps required to install and configure Harbor using the online or offline installer. The installation processes are almost identical.

If you're running an earlier version of Harbor, you may need to update the `harbor.cfg` configuration file and migrate the data to fit the new database schema. For more details, please refer to the [Harbor Migration Guide](migration_guide.md).

> **Kubernetes** — The deployment instructions for running Harbor on [Kubernetes](https://kubernetes.io) have been contributed by the Harbor community. See [Harbor on Kubernetes](kubernetes_deployment.md) for more details.

## Prerequisites for the target host

Harbor is deployed as several [Docker](https://docker.com) containers, and thus can be deployed on any Linux distribution that supports Docker. The target host requires Python, Docker, and [Docker Compose](https://docs.docker.com/compose/) to be installed.

### Hardware

The table below outlines the minimum and preferred CPU, memory, and disk for running Harbor:

Resource | Minimum | Preferred
:--------|:--------|:---------
CPU | 2 CPUs | 4 CPUs
Memory | 4 GB | 8 GB
Disk | 40 GB | 160 GB

### Software

The table below outlines the software that needs to be installed on hosts running Harbor:

Software | Version | Description
:--------|:--------|:-----------
Python | Version 2.7 or higher | Note that you may have to install Python on Linux distributions that do not come with a Python interpreter installed by default (such as Gentoo and Arch)
Docker Engine | Version 1.10 or higher | For installation instructions, see the [official documentation](https://docs.docker.com/engine/installation)
Docker Compose | Version 1.6.0 or higher | For installation instructions, see the [official documentation](https://docs.docker.com/compose/install) |
OpenSSL | Latest is preferred | Generate a certificate and keys for Harbor

### Network ports 

Harbor requires three open ports to function:

Port | Protocol | Description
:----|:---------|:-----------
443 | HTTPS | The Harbor portal and core API accept HTTPS requests on this port
4443 | HTTPS | Connections to the Docker Content Trust service for Harbor, only needed when [Notary](https://github.com/theupdateframework/notary) is enabled
80 | HTTP | The Harbor portal and core API will accept HTTP requests on this port

## Installation Steps

The installation steps boil down to the following:

1. [Download](#download-the-installer) the installer
2. [Configure](#configuring-harbor) the `harbor.cfg` configuration file
3. [Run](#finishing-installation-and-starting-harbor) `install.sh` to install and start Harbor


#### Downloading the installer

The binary for the Harbor installer can be downloaded from the [releases](https://github.com/goharbor/harbor/releases) page using a tool like [wget](https://www.gnu.org/software/wget/). Choose either the online or the offline installer. Use the `tar` command to extract the package.

Online installer:

```sh
$ tar xvf harbor-online-installer-<version>.tgz
```

Offline installer:

```sh
$ tar xvf harbor-offline-installer-<version>.tgz
```

#### Configuring Harbor

Configuration parameters are located in the `harbor.cfg` configuration file. There are two categories of parameters in `harbor.cfg`: **required parameters** and **optional parameters**.

* **Required parameters** must be set in the configuration file. They will take effect if a user updates them in `harbor.cfg` and runs the `install.sh` script to re-install Harbor.
* **Optional parameters** are optional for updating, i.e. the user can use default values and update them in the Harbor Web Portal after Harbor has been started. If they are set in `harbor.cfg` they will only take effect upon the first launch of Harbor. After that, updates to these parameters in `harbor.cfg` will be ignored. 

> **Note** — If you choose to set these parameters via the Portal, be sure to do so right after Harbor has been started. In particular, you must set the desired `auth_mode` before registering or creating any new users in Harbor. When there are users in the system (besides the default admin user), `auth_mode` cannot be changed.

The available parameters are described in the sections below. Please note that you will need to change the `hostname` attribute.

##### Required parameters

Parameter | Description | Default
:---------|:------------|:-------
`hostname` | The target host's hostname, which is used to access the Portal and the registry service. It should be the IP address or the fully qualified domain name (FQDN) of your target machine, e.g. `192.168.1.10` or `reg.yourdomain.com`. _Do NOT use `localhost` or `127.0.0.1` for the hostname - the registry service needs to be accessible by external clients!_ |
`ui_url_protocol` | Either `http` or `https`. The protocol used to access the Portal and the token/notification service.  If Notary is enabled, this parameter must be `https`; `http` is the default. To set up the HTTPS protocol, see [Configuring Harbor with HTTPS Access](configure_https.md).  | `http`
`db_password` | The root password for the PostgreSQL database used for `db_auth`. _Change this password for any production use!_ |
`max_job_workers` | The maximum number of replication workers in job service. For each image replication job, a worker synchronizes all tags of a repository to the remote destination. Increasing this number allows more concurrent replication jobs in the system. However, since each worker consumes a certain amount of network/CPU/IO resources, please carefully pick the value of this attribute based on the hardware resource of the host. | `10`
`customize_crt` | Either `on` or `off`. When this attribute is `on`, the prepare script creates a private key and root certificate for the generation/verification of the registry's token. Set this attribute to `off` when the key and root certificate are supplied by external sources. Refer to [Customize Key and Certificate of Harbor Token Service](customize_token_service.md) for more info.| `on`
`ssl_cert` | The path of the SSL certificate. This is applied only when the protocol is set to HTTPS. |
`ssl_cert_key` | The path of the SSL key. This is applied only when the protocol is set to HTTPS. |
`secretkey_path` | The path of key used to encrypt or decrypt the password of a remote registry in a replication policy. |
`log_rotate_count` | Log files are rotated `log_rotate_count` times before being removed. If the count is 0, old versions are removed rather than rotated. |
`log_rotate_size` | Log files are rotated only if they grow bigger than `log_rotate_size` bytes. If the size is followed by `k`, the size is assumed to be in kilobytes; if `M` is used, the size is in megabytes, while `G` signifies gigabytes. The sizes `100`, `100k`, `100M`, and `100G` are thus all valid. |
`http_proxy` | The HTTP proxy for Clair, e.g. `http://my.proxy.com:3128`. |
`https_proxy` | The HTTPS proxy for Clair, e.g. `http://my.proxy.com:3128`. |
`no_proxy` | Signifies no proxy for Clair, e.g. `127.0.0.1,localhost,core,registry`. |

##### Optional parameters

###### Email settings

These parameters are needed for Harbor to be able to send a user a "password reset" email and are only necessary if that functionality is needed. Also note that by default SSL connectivity is _not_ enabled. If your SMTP server requires SSL but does _not_ support STARTTLS then you should enable SSL by setting `email_ssl = true`. Set `email_insecure = true` if the email server uses a self-signed or untrusted certificate. For a detailed discussion of "email_identity" please refer to [rfc2595](https://tools.ietf.org/rfc/rfc2595.txt)

* email_server = smtp.mydomain.com
* email_server_port = 25
* email_identity = 
* email_username = sample_admin@mydomain.com
* email_password = abc
* email_from = admin <sample_admin@mydomain.com>  
* email_ssl = false
* email_insecure = false

###### Authentication and authorization settings

Parameter | Description | Default
:---------|:------------|:-------
`harbor_admin_password` | The administrator's initial password. This password only takes effect the first time that Harbor launches. After that, this setting is ignored and the administrator's password should be set in the Portal. | `Harbor12345`
`auth_mode` | The type of authentication that is used. By default, it is `db_auth`, i.e. the credentials are stored in a database. For LDAP authentication, set this to `ldap_auth`. **Important**: When upgrading from an existing Harbor instance, you must ensure that `auth_mode` is the same in `harbor.cfg` before launching the new version of Harbor. Otherwise, users may not be able to log in after the upgrade. | `db_auth`
`ldap_url` | The LDAP endpoint URL, e.g. `ldaps://ldap.mydomain.com`. _Only used when `auth_mode` is set to `ldap_auth`_. |
`ldap_searchdn` | The [DN](https://ldap.com/ldap-dns-and-rdns/) of a user who has the permission to search an LDAP/AD server, e.g. `uid=admin,ou=people,dc=mydomain,dc=com`. |
`ldap_search_pwd` | The password of the user specified by `ldap_searchdn` |
`ldap_basedn` | The base DN to look up a user, e.g. `ou=people,dc=mydomain,dc=com`. _Only used when `auth_mode` is set to `ldap_auth`_. |
`ldap_filter` | The search filter for looking up a user, e.g. `(objectClass=person)`. |
`ldap_uid` | The attribute used to match a user during a LDAP search. This could be uid, cn, email, or other attributes.|
`ldap_scope` | The scope to search for a user. Options are `0` (`LDAP_SCOPE_BASE`), `1` (`LDAP_SCOPE_ONELEVEL`), and `2` (`LDAP_SCOPE_SUBTREE`). | `2`
`ldap_timeout` | Timeout (in seconds) when connecting to an LDAP Server. | `5`
`ldap_verify_cert` | Verify certificate from LDAP server. | `true`
`ldap_group_basedn` | The base DN from which to look up a group in LDAP/AD, e.g. `ou=group,dc=mydomain,dc=com`. |
`ldap_group_filter` | The filter used to search the LDAP/AD group, e.g. `objectclass=group`. |
`ldap_group_gid` | The attribute used to name the LDAP/AD group. This could be cn or name. |
`ldap_group_scope` | The scope to search for ldap groups. Options are `0` (`LDAP_SCOPE_BASE`), `1` (`LDAP_SCOPE_ONELEVEL`), and `2` (`LDAP_SCOPE_SUBTREE`) | `2`
`self_registration` | Either `on` or `off`. Enables/disables the ability for a user to register themselves. When disabled, new users can only be created by the Admin user (only an admin user can create new users in Harbor). **Note**: When `auth_mode` is set to `ldap_auth`, the self-registration feature is *always* disabled, and this flag is ignored. | `on`
`token_expiration` | The expiration time (in minutes) of a token created by the token service. | `30`
`project_creation_restriction` | The flag used to control which users have the permission to create projects. By default everyone can create a project. Set this to `adminonly` to ensure that only admins can create projects. |

#### Configuring storage backend (optional)

By default, Harbor stores images on your local filesystem. In a production environment, you may consider 
using other storage backend instead of the local filesystem, like S3, OpenStack Swift, Ceph, etc.
These parameters are configurations for registry.

* **registry_storage_provider_name**:  Storage provider name of registry, it can be filesystem, s3, gcs, azure, etc. Default is filesystem.
* **registry_storage_provider_config**: Comma separated "key: value" pairs for storage provider config, e.g. "key1: value, key2: value2". Default is empty string.
* **registry_custom_ca_bundle**:  The path to the custom root ca certificate, which will be injected into the truststore of registry's and chart repository's containers.  This is usually needed when the user hosts a internal storage with self signed certificate.

For example, if you use Openstack Swift as your storage backend, the parameters may look like this:

```ini
registry_storage_provider_name=swift
registry_storage_provider_config="username: admin, password: ADMIN_PASS, authurl: http://keystone_addr:35357/v3/auth, tenant: admin, domain: default, region: regionOne, container: docker_images"
```

_NOTE: For detailed information on storage backend of a registry, refer to [Registry Configuration Reference](https://docs.docker.com/registry/configuration/) ._


#### Finishing installation and starting Harbor
Once **harbor.cfg** and storage backend (optional) are configured, install and start Harbor using the `install.sh` script.  Note that it may take some time for the online installer to download Harbor images from Docker hub.  

##### Default installation (without Notary/Clair)
Harbor has integrated with Notary and Clair (for vulnerability scanning). However, the default installation does not include Notary or Clair service.

```sh
$ sudo ./install.sh
```

If everything worked properly, you should be able to open a browser to visit the admin portal at `http://reg.yourdomain.com` (change `reg.yourdomain.com` to the hostname configured in your `harbor.cfg`). Note that the default administrator username/password is `admin`/`Harbor12345`.

Log in to the admin portal and create a new project, e.g. `myproject`. You can then use Docker CLI commands to log in and push images (the registry server listens on port 80 by default):

```sh
$ docker login reg.yourdomain.com
$ docker push reg.yourdomain.com/myproject/myrepo:mytag
```

> **Important** — The default installation of Harbor uses HTTP. Thus, you'll need to add the `--insecure-registry` option to your client's Docker daemon and restart the Docker service. 

##### Installation with Notary

To install Harbor with the Notary service, add the `--with-notary` flag when you run `install.sh`:

```sh
$ sudo ./install.sh --with-notary
```

> **Note** — For installation with Notary, the `ui_url_protocol` parameter must be set to `https`. For configuring HTTPS please refer to the sections below.

More information about Notary and Docker Content Trust, see the [Docker documentation](https://docs.docker.com/engine/security/trust/content_trust).

##### Installation with Clair

To install Harbor with the Clair service, add the `--with-clair` flag when you run `install.sh`:

```sh
$ sudo ./install.sh --with-clair
```

For more information on Clair, see the [Clair documentation](https://coreos.com/clair/docs/2.0.1).

##### Installation with the chart repository service

To install Harbor with the chart repository service, add the `--with-chartmuseum` flag when you run `install.sh`:

```sh
$ sudo ./install.sh --with-chartmuseum
```

> **Note** — If you want to install Notary, Clair and chart repository service, you must specify all the parameters in the same command:

```sh
$ sudo ./install.sh --with-notary --with-clair --with-chartmuseum
```

For information on using Harbor, see the **[Harbor User Guide](user_guide.md)** .

#### Configuring Harbor with HTTPS access

Harbor does not ship with any certificates, and, by default, uses HTTP to serve requests. While this makes it relatively simple to set up and run - especially for a development or testing environment - it is **not** recommended for a production environment.  To enable HTTPS, please refer to **[Configuring Harbor with HTTPS Access](configure_https.md)**.  


### Managing Harbor's lifecycle

You can use docker-compose to manage the lifecycle of Harbor. Some useful commands are listed as follows (must run in the same directory as *docker-compose.yml*).

Stopping Harbor:

```sh
$ sudo docker-compose stop
Stopping nginx              ... done
Stopping harbor-portal      ... done
Stopping harbor-jobservice  ... done
Stopping harbor-core        ... done
Stopping registry           ... done
Stopping redis              ... done
Stopping registryctl        ... done
Stopping harbor-db          ... done
Stopping harbor-adminserver ... done
Stopping harbor-log         ... done
```  

Restarting Harbor after stopping:

```sh
$ sudo docker-compose start
Starting log         ... done
Starting registry    ... done
Starting registryctl ... done
Starting postgresql  ... done
Starting adminserver ... done
Starting core        ... done
Starting portal      ... done
Starting redis       ... done
Starting jobservice  ... done
Starting proxy       ... done
```  

To change Harbor's configuration, first stop existing Harbor instance and update `harbor.cfg`. Then run `prepare` script to populate the configuration. Finally re-create and start Harbor's instance:

```sh
$ sudo docker-compose down -v
$ vim harbor.cfg
$ sudo prepare
$ sudo docker-compose up -d
``` 

Removing Harbor's containers while keeping the image data and Harbor's database files on the file system:

```sh
$ sudo docker-compose down -v
```  

Removing Harbor's database and image data (for a clean re-installation):

```sh
$ rm -r /data/database
$ rm -r /data/registry
```

#### _Managing lifecycle of Harbor when it's installed with Notary_ 

When Harbor is installed with Notary, an extra template file `docker-compose.notary.yml` is needed for docker-compose commands. The docker-compose commands to manage the lifecycle of Harbor are:

```sh
$ sudo docker-compose -f ./docker-compose.yml -f ./docker-compose.notary.yml [ up|down|ps|stop|start ]
```

For example, if you want to change configuration in `harbor.cfg` and re-deploy Harbor when it's installed with Notary, the following commands should be used:

```sh
$ sudo docker-compose -f ./docker-compose.yml -f ./docker-compose.notary.yml down -v
$ vim harbor.cfg
$ sudo prepare --with-notary
$ sudo docker-compose -f ./docker-compose.yml -f ./docker-compose.notary.yml up -d
```

#### _Managing lifecycle of Harbor when it's installed with Clair_ 

When Harbor is installed with Clair, an extra template file called `docker-compose.clair.yml` is needed for `docker-compose` commands. Here are the commands necessary to manage the lifecycle of Harbor:

```sh
$ sudo docker-compose -f ./docker-compose.yml -f ./docker-compose.clair.yml [ up|down|ps|stop|start ]
```

For example, if you want to change configuration in `harbor.cfg` and re-deploy Harbor when it's installed with Clair, the following commands should be used:

```sh
$ sudo docker-compose -f ./docker-compose.yml -f ./docker-compose.clair.yml down -v
$ vim harbor.cfg
$ sudo prepare --with-clair
$ sudo docker-compose -f ./docker-compose.yml -f ./docker-compose.clair.yml up -d
```

#### _Managing lifecycle of Harbor when it's installed with chart repository service_ 

When Harbor is installed with chart repository service, an extra template file `docker-compose.chartmuseum.yml` is needed for docker-compose commands. The docker-compose commands to manage the lifecycle of Harbor are:

```sh
$ sudo docker-compose -f ./docker-compose.yml -f ./docker-compose.chartmuseum.yml [ up|down|ps|stop|start ]
```

For example, if you want to change configuration in `harbor.cfg` and re-deploy Harbor when it's installed with chart repository service, the following commands should be used:

```sh
$ sudo docker-compose -f ./docker-compose.yml -f ./docker-compose.chartmuseum.yml down -v
$ vim harbor.cfg
$ sudo prepare --with-chartmuseum
$ sudo docker-compose -f ./docker-compose.yml -f ./docker-compose.chartmuseum.yml up -d
```

#### _Managing lifecycle of Harbor when it's installed with Notary, Clair and chart repository service_ 

If you want to install Notary, Clair and chart repository service together, you should include all the components in the docker-compose and prepare commands:

```sh
$ sudo docker-compose -f ./docker-compose.yml -f ./docker-compose.notary.yml -f ./docker-compose.clair.yml -f ./docker-compose.chartmuseum.yml down -v
$ vim harbor.cfg
$ sudo prepare --with-notary --with-clair --with-chartmuseum
$ sudo docker-compose -f ./docker-compose.yml -f ./docker-compose.notary.yml -f ./docker-compose.clair.yml -f ./docker-compose.chartmuseum.yml up -d
```

Please check the [Docker Compose command-line reference](https://docs.docker.com/compose/reference/) for more on docker-compose.

### Persistent data and log files

By default, registry data is persisted in the host's `/data/` directory.  This data remains unchanged even when Harbor's containers are removed and/or recreated.  

In addition, Harbor uses [rsyslog](https://www.rsyslog.com/) to collect the logs from each container. By default, these log files are stored in the `/var/log/harbor/` directory on the target host for troubleshooting.  

## Configuring Harbor to listen on a customized port

By default, Harbor listens on ports 80 (HTTP) and 443 (HTTPS, if configured) for both admin portal and Docker commands. But you can also configure it to listen on custom ports. 

### HTTP

1. Modify `docker-compose.yml` and replace the first `80` with a custom port, e.g. `8888:80`.

    ```yaml
    proxy:
    image: goharbor/nginx-photon:v1.6.0
    container_name: nginx
    restart: always
    volumes:
    - ./common/config/nginx:/etc/nginx:z
    ports:
    - 8888:80
    - 443:443
    depends_on:
    - postgresql
    - registry
    - core
    - portal
    - log
    logging:
        driver: "syslog"
        options:
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "proxy"
    ```

2.Modify `harbor.cfg`, adding the port to the `hostname` parameter:

    ```conf
    hostname = 192.168.0.2:8888
    ```

3. Re-deploy Harbor using the instructions in the [Managing Harbor's Lifecycle](#managing-harbors-lifecycle) section.

### For HTTPS protocol

1.Enable HTTPS in Harbor by following the [Configuring Harbor with HTTPS Access](https://github.com/goharbor/harbor/blob/master/docs/configure_https.md) guide.

2. Modify `docker-compose.yml`, replacing the `443` with a customized port, e.g. `8888:443`.

    ```yaml
    proxy:
    image: goharbor/nginx-photon:v1.6.0
    container_name: nginx
    restart: always
    volumes:
    - ./common/config/nginx:/etc/nginx:z
    ports:
    - 80:80
    - 8888:443
    depends_on:
    - postgresql
    - registry
    - core
    - portal
    - log
    logging:
        driver: "syslog"
        options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "proxy"
    ```

3. Modify `harbor.cfg`, adding the port to the `hostname` parameter:

    ```conf
    hostname = 192.168.0.2:8888
    ```

4. Re-deploy Harbor using the instructions in the [Managing Harbor's Lifecycle](#managing-harbors-lifecycle) section.

## Performance tuning

By default, Harbor limits the CPU usage of the Clair container to 150000 and prevents it from using excessive CPU resources. This is defined in the `docker-compose.clair.yml` file. You can modify it based on your hardware configuration.

## Troubleshooting

1. When Harbor isn't working properly, run the commands below to find out if all containers related to Harbor have the `UP` status:

    ```sh
    $ sudo docker-compose ps
            Name                     Command               State                    Ports                   
    -----------------------------------------------------------------------------------------------------------------------------
    harbor-adminserver  /harbor/start.sh                 Up
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

    If a container is not in the `UP` state, check the log file of that container in the `/var/log/harbor` directory. If the container `harbor-core` is not running, for example, you should check the `core.log` log file.


2. When setting up Harbor behind an nginx proxy or elastic load balancing, look for the line below, in `common/templates/nginx/nginx.http.conf` and remove it from the sections if the proxy already has similar settings: `location /`, `location /v2/` and `location /service/`.

    ```conf
    proxy_set_header X-Forwarded-Proto $scheme;
    ```

3. Re-deploy Harbor using the instructions in the [Managing Harbor's Lifecycle](#managing-harbors-lifecycle) section.
