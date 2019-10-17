# Configure the Harbor YML File

Configuration parameters are located in the file **harbor.yml**.

There are two categories of parameters, **required parameters** and **optional parameters**.

- **System level parameters**: These parameters are required to be set in the configuration file. They will take effect if a user updates them in ```harbor.yml``` and run the ```install.sh``` script to reinstall Harbor.

- **User level parameters**: These parameters can update after the first time harbor started on Web Portal. In particular, you must set the desired **auth_mode** before registering or creating any new users in Harbor. When there are users in the system (besides the default admin user), **auth_mode** cannot be changed.

The parameters are described below - note that at the very least, you will need to change the **hostname** attribute.

##### Required parameters

- **hostname**: The target host's hostname, which is used to access the Portal and the registry service. It should be the IP address or the fully qualified domain name (FQDN) of your target machine, e.g., `192.168.1.10` or `reg.yourdomain.com`. _Do NOT use `localhost` or `127.0.0.1` or `0.0.0.0` for the hostname - the registry service needs to be accessible by external clients!_

- **data_volume**: The location to store harbor's data.

- **harbor_admin_password**: The administrator's initial password. This password only takes effect for the first time Harbor launches. After that, this setting is ignored and the administrator's password should be set in the Portal. _Note that the default username/password are **admin/Harbor12345** ._

- **database**: the configs related to local database
  - **password**: The root password for the PostgreSQL database. Change this password for any production use.
  - **max_idle_conns**: The maximum number of connections in the idle connection pool. If <=0 no idle connections are retained. The default value is 50 and if it is not configured the value is 2.
  - **max_open_conns**: The maximum number of open connections to the database. If <= 0 there is no limit on the number of open connections. The default value is 100 for the max connections to the Harbor database. If it is not configured the value is 0.

- **jobservice**: jobservice related service
  - **max_job_workers**: The maximum number of replication workers in job service. For each image replication job, a worker synchronizes all tags of a repository to the remote destination. Increasing this number allows more concurrent replication jobs in the system. However, since each worker consumes a certain amount of network/CPU/IO resources, please carefully pick the value of this attribute based on the hardware resource of the host.
- **log**: log related url
  - **level**: log level, options are debug, info, warning, error, fatal
  - **local**: The default is to retain logs locally.
      - **rotate_count**: Log files are rotated **rotate_count** times before being removed. If count is 0, old versions are removed rather than rotated.
      - **rotate_size**: Log files are rotated only if they grow bigger than **rotate_size** bytes. If size is followed by k, the size is assumed to be in kilobytes. If the M is used, the size is in megabytes, and if G is used, the size is in gigabytes. So size 100, size 100k, size 100M and size 100G are all valid.
      - **location**: the directory to store logs
  - **external_endpoint**: Enable this option to forward logs to a syslog server.
       - **protocol**: Transport protocol for the syslog server. Default is TCP.
       - **host**: The URL of the syslog server.
       - **port**: The port on which the syslog server listens.
     
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
    - **max_idle_conns**: The maximum number of connections in the idle connection pool. If <=0 no idle connections are retained. The default value  is 2.
    - **max_open_conns**: The maximum number of open connections to the database. If <= 0 there is no limit on the number of open connections. The default value is 0.
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

# Configuring Harbor listening on a customized port

By default, Harbor listens on port 80(HTTP) and 443(HTTPS, if configured) for both admin portal and docker commands, these default ports can configured in `harbor.yml`

## Configuring Harbor using the external database

Currently, only PostgreSQL database is supported by Harbor.
To user an external database, just uncomment the `external_database` section in `harbor.yml` and fill the necessary information. Four databases are needed to be create first by users for Harbor core, Clair, Notary server and Notary signer. And the tables will be generated automatically when Harbor starting up.

## Manage user settings

After release 1.8.0, User settings are separated with system settings, and all user settings should be configured in web console or by HTTP request.
Please refer [Configure User Settings](configure_user_settings.md) to config user settings.

## Performance tuning

By default, Harbor limits the CPU usage of Clair container to 150000 and avoids its using up all the CPU resources. This is defined in the docker-compose.clair.yml file. You can modify it based on your hardware configuration.