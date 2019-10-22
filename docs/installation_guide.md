# Installation and Configuration Guide

There are two possibilities when installing Harbor.

- **Online installer:** The online installer downloads the Harbor images from Docker hub. For this reason, the installer is very small in size.

- **Offline installer:** Use the offline installer if the host to which are are deploying Harbor does not have a connection to the Internet. The offline installer contains pre-built images so it is larger than the online installer.

You download the installers from the **[official release](https://github.com/goharbor/harbor/releases)** page.

This guide describes how to install and configure Harbor by using either the online or offline installer. The installation processes are almost the same.

If you are upgrading from a previous version of Harbor, you might need to update the configuration file and migrate your data to fit the database schema of the later version. For information about upgrading, see the **[Harbor Upgrade and Migration Guide](migration_guide.md)**.

In addition, the Harbor community created instructions describing how to deploy Harbor on Kubernetes. If you want to deploy Harbor to Kubernetes, see [Harbor on Kubernetes](kubernetes_deployment.md).

## Harbor Components

The table below lists the components that are deployed when you deploy Harbor.

|Component|Version|
|---|---|
|Postgresql|9.6.10-1.ph2|
|Redis|4.0.10-1.ph2|
|Clair|2.0.8|
|Beego|1.9.0|
|Chartmuseum|0.9.0|
|Docker/distribution|2.7.1|
|Docker/notary|0.6.1|
|Helm|2.9.1|
|Swagger-ui|3.22.1|

## Deployment Prerequisites for the Target Host

Harbor is deployed as several Docker containers. You can therefore deploy it on any Linux distribution that supports Docker. The target host requires Docker, and Docker Compose to be installed.

### Hardware

The following table lists the minimum and recommended hardware configurations for deploying Harbor.

|Resource|Minimum|Recommended|
|---|---|---|
|CPU|2 CPU|4 CPU|
|Mem|4 GB|8 GB|
|Disk|40 GB|160 GB|

### Software

The following table lists the software versions that must be installed on the target host.

|Software|Version|Description|
|---|---|---|
|Docker engine|version 17.06.0-ce+ or higher|For installation instructions, see [docker engine doc](https://docs.docker.com/engine/installation/)|
|Docker Compose|version 1.18.0 or higher|For installation instructions, see [docker compose doc](https://docs.docker.com/compose/install/)|
|Openssl|latest is preferred|Used to generate certificate and keys for Harbor|

### Network ports

Harbor requires that the following ports be open on the target host.

|Port|Protocol|Description|
|---|---|---|
|443|HTTPS|Harbor portal and core API accept HTTPS requests on this port. You can change this port in the configuration file.|
|4443|HTTPS|Connections to the Docker Content Trust service for Harbor. Only required if Notary is enabled. You can change this port in the configuration file.|
|80|HTTP|Harbor portal and core API accept HTTP requests on this port. You can change this port in the configuration file.|

## Installation Procedure

The installation procedure involves the following steps:

1. Download the installer.
2. Configure the **harbor.yml** file.
3. Run the **install.sh** script with the appropriate options to install and start Harbor.

## Download the Installer

1. Go to the [Harbor releases page](https://github.com/goharbor/harbor/releases). 
1. Select either the online or offline installer for the version you want to install.
1. Use `tar` to extract the installer package:

   - Online installer:<pre>bash $ tar xvf harbor-online-installer-<em>version</em>.tgz</pre>
   - Offline installer:<pre>bash $ tar xvf harbor-offline-installer-<em>version</em>.tgz</pre>

## Configure Harbor

You set system level parameters for Harbor in the `harbor.yml` file that is contained in the installer package. These parameters take effect when you run the `install.sh` script to install or reconfigure Harbor. 

After the initial deployment and after you have started Harbor, you perform additional configuration in the Harbor Web Portal. 

### Required Parameters

The table below lists the parameters that must be set when you deploy Harbor. By default, all of the required parameters are uncommented in the `harbor.yml` file. The optional parameters are commented with `#`. You do not necessarily need to change the values of the required parameters from the defaults that are provided, but these parameters must remain uncommented. At the very least, you must update the `hostname` parameter.

**IMPORTANT**: Harbor does not ship with any certificates, and by default uses HTTP to serve registry requests. This is acceptable only in air-gapped test or development environments. In production environments, always use HTTPS. If you enable Content Trust with Notary to properly sign all images, you must use HTTPS. 
  
You can use certificates that are signed by a trusted third-party CA, or you can use self-signed certificates. For information about how to create a CA, and how to use a CA to sign a server certificate and a client certificate, see **[Configuring Harbor with HTTPS Access](configure_https.md)**.

<table width="100%" border="0">
  <caption>
    Required Parameters for Harbor Deployment
  </caption>
  <tr>
    <th scope="col">Parameter</th>
    <th scope="col">Sub-parameters</th>
    <th scope="col">Description and Additional Parameters </th>
  </tr>
  <tr>
    <td valign="top"><code>hostname</code></td>
    <td valign="top">None</td>
    <td valign="top">Specify the IP address or the fully qualified domain name (FQDN) of the target host on which to deploy Harbor. This is the address at which you access the Harbor Portal and the registry service. For example, <code>192.168.1.10</code> or <code>reg.yourdomain.com</code>. The registry service must be accessible to external clients, so do not specify <code>localhost</code>, <code>127.0.0.1</code>, or <code>0.0.0.0</code> as the hostname.</td>
  </tr>
  <tr>
    <td valign="top"><code>https</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top"><p>Use HTTPS to access the Harbor Portal and the token/notification service. </p>
      </td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>port</code></td>
    <td valign="top">The port number for HTTPS. The default is 443.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>certificate</code></td>
    <td valign="top">The path to the SSL certificate.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>private_key</code></td>
    <td valign="top">The path to the SSL key.</td>
  </tr>
  <tr>
    <td valign="top"><code>harbor_admin_password</code></td>
    <td valign="top">None</td>
    <td valign="top">Set an initial password for the Harbor administrator. This password is only used on the first time that Harbor starts. On subsequent logins, this setting is ignored and the administrator's password is set in the Harbor Portal. The default username and password are <code>admin</code> and <code>Harbor12345</code>.</td>
  </tr>
  <tr>
    <td valign="top"><code>database</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Use a local PostgreSQL database. You can optionally configure an external database, in which case you can disable this option.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>password</code></td>
    <td valign="top">Set the root password for the local database. You must change this password for production deployments.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>max_idle_conns</code></td>
    <td valign="top">The maximum number of connections in the idle connection pool. If set to &lt;=0 no idle connections are retained. The default value is 50. If it is not configured the value is 2.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>max_open_conns</code></td>
    <td valign="top">The maximum number of open connections to the database. If &lt;= 0 there is no limit on the number of open connections. The default value is 100 for the max connections to the Harbor database. If it is not configured the value is 0.</td>
  </tr>
  <tr>
    <td valign="top"><code>data_volume</code></td>
    <td valign="top">None</td>
    <td valign="top">The location on the target host in which to store Harbor's data. You can optionally configure external storage, in which case disable this option and enable <code>storage_service</code>. The default is <code>/data</code>.</td>
  </tr>
  <tr>
    <td valign="top"><code>clair</code></td>
    <td valign="top"><code>updaters_interval</code></td>
    <td valign="top">Set an interval for Clair updates, in hours. Set to 0 to disable the updates. The default is 12 hours.</td>
  </tr>
  <tr>
    <td valign="top"><code>jobservice</code></td>
    <td valign="top"><code>max_job_workers</code></td>
    <td valign="top">The maximum number of replication workers in the job service. For each image replication job, a worker synchronizes all tags of a repository to the remote destination. Increasing this number allows more concurrent replication jobs in the system. However, since each worker consumes a certain amount of network/CPU/IO resources, set the value of this attribute based on the hardware resource of the host. The default is 10.</td>
  </tr>
<tr>
    <td valign="top"><code>notification</code></td>
    <td valign="top"><code>webhook_job_max_retry</code></td>
    <td valign="top">Set the maximum number of retries for web hook jobs. The default is 10.</td>
  </tr>
  <tr>
    <td valign="top"><code>chart</code></td>
    <td valign="top"><code>absolute_url</code></td>
    <td valign="top">Set to <code>enabled</code> for Chart to use an absolute URL. Set to <code>disabled</code> for Chart to use a relative URL.</td>
  </tr>
  <tr>
    <td valign="top"><code>log</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Configure logging.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>level</code></td>
    <td valign="top">Set the logging level to <code>debug</code>, <code>info</code>, <code>warning</code>, <code>error</code>, or <code>fatal</code>. The default is <code>info</code>.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>local</code></td>
    <td valign="top">Set the log retention parameters:<ul>
          <li><code>rotate_count</code>: Log files are rotated <code>rotate_count</code> times before being removed. If count is 0, old versions are removed rather than rotated. The default is 50.</li>
          <li><code>rotate_size</code>: Log files are rotated only if they grow bigger than <code>rotate_size</code> bytes. Use <code>k</code> for kilobytes, <code>M</code> for megabytes, and <code>G</code> for gigabytes.  <code>100</code>, <code>100k</code>, <code>100M</code> and <code>100G</code> are all valid values. The default is 200M.</li>
          <li><code>location</code>: Set the directory in which to store the logs. The default is <code>/var/log/harbor</code>.</li>
        </ul></td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>external_endpoint</code></td>
    <td valign="top">Enable this option to forward logs to a syslog server.
      <ul>
        <li><code>protocol</code>: Transport protocol for the syslog server. Default is TCP.</li>
        <li><code>host</code>: The URL of the syslog server.</li>
        <li><code>port</code>: The port on which the syslog server listens</li>
    </ul>    </td>
  </tr>
  <tr>
    <td valign="top"><code>proxy</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Configure proxies to be used by Clair, the replication jobservice, and Harbor.</td>
  </tr>
    <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>http_proxy</code></td>
    <td valign="top">Configure an HTTP proxy, for example,  <code>http://my.proxy.com:3128</code>.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>https_proxy</code></td>
    <td valign="top">Configure an HTTPS proxy, for example,  <code>http://my.proxy.com:3128</code>.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>no_proxy</code></td>
    <td valign="top">Configure when not to use a proxy, for example, <code>127.0.0.1,localhost,core,registry</code>.</td>
  </tr>
</table>
  
### Optional parameters

The following table lists the additional, optional parameters that you can set to configure your Harbor deployment beyond the minimum required settings. To enable a setting, you must uncomment it in `harbor.yml` by deleting the leading `#` character.

<table width="100%" border="0">
  <caption>
    Optional Parameters for Harbor
  </caption>
  <tr>
    <th scope="col">Parameter</th>
    <th scope="col">Sub-Parameters</th>
    <th scope="col">Description and Additional Parameters </th>
  </tr>
  <tr>
    <td valign="top"><code>http</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Do not use HTTP in production environments. Using HTTP is acceptable only in air-gapped test or development environments that do not have a connection to the external internet. Using HTTP in environments that are not air-gapped exposes you to man-in-the-middle attacks.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>port</code></td>
    <td valign="top">Port number for HTTP</td>
  </tr>
  <tr>
    <td valign="top"><code>external_url</code></td>
    <td valign="top">None</td>
    <td valign="top">Enable this option to use an external proxy. When  enabled, the hostname is no longer used.</td>
  </tr>
  <tr>
  <tr>
    <td valign="top"><code>storage_service</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">By default, Harbor stores images and charts on your local filesystem. In a production environment, you might want to use another storage backend instead of the local filesystem. The parameters listed below are the configurations for the registry. See *Configuring Storage Backend* below for more information about how to configure a different backend.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>ca_bundle</code></td>
    <td valign="top">The path to the custom root CA certificate, which is injected into the trust store of registry and chart repository containers. This is usually needed if internal storage uses a self signed certificate.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>filesystem</code></td>
    <td valign="top">The default is <code>filesystem</code>, but you can set <code>azure</code>, <code>gcs</code>, <code>s3</code>, <code>swift</code> and <code>oss</code>. For information about how to configure other backends, see <a href="#backend">Configuring a Storage Backend</a> below. Set <code>maxthreads</code> to limit the number of threads to the external provider. The default is 100.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>redirect</code></td>
    <td valign="top">Set <code>disable</code> to <code>true</code> when you want to disable registry redirect</td>
  </tr>
  <tr>
    <td valign="top"><code>external_database</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Configure external database settings, if you disable the local database option. Harbor currently only supports POSTGRES.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>harbor</code></td>
    <td valign="top"><p>Configure an external database for Harbor data.</p>
      <ul>
        <li><code>host</code>: Hostname of the Harbor database.</li>
        <li><code>port</code>: Database port.</li>
        <li><code>db_name</code>: Database name.</li>
        <li><code>username</code>: Username to connect to the core Harbor database.</li>
        <li><code>password</code>: Password for the account you set in <code>username</code>.</li>
        <li><code>ssl_mode</code>: Enable SSL mode.</li>
        <li><code>max_idle_conns</code>: The maximum number of connections in the idle connection pool. If &lt;=0 no idle connections are retained. The default value is 2.</li>
        <li><code>max_open_conns</code>: The maximum number of open connections to the database. If &lt;= 0 there is no limit on the number of open connections. The default value is 0.</li>
    </ul>      </td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>clair</code></td>
    <td valign="top">Configure an external database for Clair.
      <ul>
        <li><code>host</code>: Hostname of the Clair database</li>
        <li><code>port</code>: Database port.</li>
        <li><code>db_name</code>: Database name.</li>
        <li><code>username</code>: Username to connect to the Clair database.</li>
        <li><code>password</code>: Password for the account you set in <code>username</code>.</li>
        <li><code>ssl_mode</code>: Enable SSL mode.</li>
      </ul>    </td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>notary_signer</code></td>
    <td valign="top">Configure an external database for the Notary signer database
      <ul>
        <li><code>host</code>: Hostname of the Notary signer database</li>
        <li><code>port</code>: Database port.</li>
        <li><code>db_name</code>: Database name.</li>
        <li><code>username</code>: Username to connect to the Notary signer database.</li>
        <li><code>password</code>: Password for the account you set in <code>username</code>.</li>
        <li><code>ssl_mode</code>: Enable SSL mode.</li>
      </ul>    </td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>notary_server</code></td>
    <td valign="top"><ul>
      <li><code>host</code>: Hostname of the Notary server database.</li>
      <li><code>port</code>: Database port.</li>
      <li><code>db_name</code>: Database name.</li>
      <li><code>username</code>: Username to connect to the Notary server database.</li>
      <li><code>password</code>: Password for the account you set in <code>username</code>.</li>
      <li><code>ssl_mode</code>: Enable SSL mode.e</li>
    </ul>    </td>
  </tr>
  <tr>
    <td valign="top"><code>external_redis</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Configure an external Redis instance.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>host</code></td>
    <td valign="top">Hostname of the external Redis instance.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>port</code></td>
    <td valign="top">Redis instance port.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>password</code></td>
    <td valign="top">Password to connect to the external Redis instance.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>registry_db_index</code></td>
    <td valign="top">Database index for Harbor registry.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>jobservice_db_index</code></td>
    <td valign="top">Database index for jobservice.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>chartmuseum_db_index</code></td>
    <td valign="top">Database index for Chart museum.</td>
  </tr>
  <tr>
    <td valign="top"><code>uaa</code></td>
    <td valign="top">&nbsp;</td>
    <td valign="top">Enable UAA to trust the certificate of a UAA instance that is hosted via a self-signed certificate.</td>
  </tr>
  <tr>
    <td valign="top">&nbsp;</td>
    <td valign="top"><code>ca_file</code></td>
    <td valign="top">The path to the self-signed certificate of the UAA instance, for example <code>/path/to/ca</code>.</td>
  </tr>
</table>

<a id="backend"></a>
### Configuring a Storage Backend 

By default Harbor uses local storage for the registry, but you can optionally configure the `storage_service` setting so that Harbor uses external storage. For information about how to configure the storage backend of a registry for different storage providers, see the [Registry Configuration Reference](https://docs.docker.com/registry/configuration/#storage) in the Docker documentation. For example, if you use Openstack Swift as your storage backend, the parameters might resemble the following:

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


## Installating and starting Harbor

Once you have configured **harbor.yml** optionally set up a storage backend, you install and start Harbor by using the `install.sh` script. Note that it might take some time for the online installer to download all of the `Harbor images from Docker hub.

You can install Harbor in different configurations:

- Just Harbor, without Notary, Clair, or Chart Repository Service
- Harbor with Notary
- Harbor with Clair
- Harbor with Chart Repository Service
- Harbor with two or all three of Notary, Clair, and Chart Repository Service

### Default installation without Notary, Clair, or Chart Repository Service

The default Harbor installation does not include Notary or Clair service.

``` sh
    $ sudo ./install.sh
```

If the installation succeeds, you can open a browser to visit the Harbor Portal at `http://reg.yourdomain.com`, changing `reg.yourdomain.com` to the hostname that you configured in `harbor.yml`. If you did not change them, the default administrator username and password are `admin` and `Harbor12345`.

Log in to the admin portal and create a new project, for example, `myproject`. You can then use docker commands to log in and push images to Harbor. By default, the registry server listens on port 80:

```sh
$ docker login reg.yourdomain.com
$ docker push reg.yourdomain.com/myproject/myrepo:mytag
```

**IMPORTANT:** If your installation of Harbor uses HTTP, you must add the option `--insecure-registry` to your client's Docker daemon and restart the Docker service.

### Installation with Notary

To install Harbor with the Notary service, add the `--with-notary` parameter when you run `install.sh`:

```sh
    $ sudo ./install.sh --with-notary
```

**Note**: For installation with Notary, you must use Harbor with HTTPS.

For more information about Notary and Docker Content Trust, see [Content Trust](https://docs.docker.com/engine/security/trust/content_trust/) in the Docker documentation.

### Installation with Clair

To install Harbor with Clair service, add the `--with-clair` parameter when you run `install.sh`:

```sh
    $ sudo ./install.sh --with-clair
```

For more information about Clair, see the [Clair documentation](https://coreos.com/clair/docs/2.0.1/).

### Installation with Chart Repository Service 

To install Harbor with chart repository service, add the `--with-chartmuseum` parameter when you run ```install.sh```:

```sh
    $ sudo ./install.sh --with-chartmuseum
```

### Installation with Notary, Clair, and Chart Repository Service

If you want to install all three of Notary, Clair and chart repository service, you must specify all of the parameters in the same command:

```sh
    $ sudo ./install.sh --with-notary --with-clair --with-chartmuseum
```

## Using Harbor

For information on how to use Harbor, see the **[Harbor User Guide](user_guide.md)** .

## Managing Harbor's lifecycle

You can use `docker-compose` to manage the lifecycle of Harbor. Some useful commands are listed below. You must run the commands in the same directory as `docker-compose.yml`.

### Stop Harbor:

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

### Restart Harbor after Stopping:

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

### Reconfigure Harbor

To reconfigure Harbor, stop the existing Harbor instance and update `harbor.yml`. Then run `prepare` script to populate the configuration. Finally re-create and start the Harbor instance.

``` sh
$ sudo docker-compose down -v
$ vim harbor.yml
$ sudo prepare
$ sudo docker-compose up -d
```

### Other Commands

Remove Harbor's containers while keeping the image data and Harbor's database files on the file system:

``` sh
$ sudo docker-compose down -v
```

Remove Harbor's database and image data for a clean re-installation:

``` sh
$ rm -r /data/database
$ rm -r /data/registry
```

### Managing the Harbor Lifecycle  with Notary, Clair and Chart Repository Service

If you want to install Notary, Clair and chart repository service together, you should include all the components in the prepare commands:

``` sh
$ sudo docker-compose down -v
$ vim harbor.yml
$ sudo prepare --with-notary --with-clair --with-chartmuseum
$ sudo docker-compose up -d
```

Please check the [Docker Compose command-line reference](https://docs.docker.com/compose/reference/) for more on docker-compose.

## Persistent Data and Log Files

By default, registry data is persisted in the host's `/data/` directory.  This data remains unchanged even when Harbor's containers are removed and/or recreated. You can edit the `data_volume` in `harbor.yml` file to change this directory.

In addition, Harbor uses `rsyslog` to collect the logs for each container. By default, these log files are stored in the directory `/var/log/harbor/` on the target host. You can change the log directory in `harbor.yml`.

## Configuring Harbor to Listen on a Customized Port

By default, Harbor listens on port 443(HTTPS) and 80(HTTP, if configured)  for both Harbor portal and Docker commands. You can reconfigure the default ports in `harbor.yml`

## Configure Harbor with an External Database

Currently, Harbor only supports PostgreSQL database. To user an external database, uncomment the `external_database` section in `harbor.yml` and fill the necessary information. You must create four databases for Harbor core, Clair, Notary server, and Notary signer. And the tables are generated automatically when Harbor starts up.

## Manage User Settings

User settings are handled separately system settings. All user settings are configured in the Harbor portal or by HTTP requests at the command line. For information about using HTTP requests to configure user settings, see  [Configure User Settings at the Command Line](configure_user_settings.md) to config user settings.

## Performance Tuning

By default, Harbor limits the CPU usage of the Clair container to 150000 to avoid it using up all CPU resources. This is defined in the `docker-compose.clair.yml` file. You can modify this file based on your hardware configuration.

## Troubleshooting

### Harbor Doesn't Start or Functions Incorrectly

When Harbor does not function correctly, run the following commands to find out if all of Harbor's containers in **UP** status:
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

If a container is not in the `Up` state, check the log file for that container in `/var/log/harbor`. For example, if the `harbor-core` container is not running, look at the `core.log` log file.

### Using nginx or Load Balancing

When setting up Harbor behind an `nginx` proxy or elastic load balancing, look for the following line in `common/config/nginx/nginx.conf` and, if the proxy already has similar settings, remove it from the sections `location /`, `location /v2/` and `location /service/`.

``` sh
proxy_set_header X-Forwarded-Proto $scheme;
```

Then re-deploy Harbor per the instructions in "Managing Harbor Lifecycle.
