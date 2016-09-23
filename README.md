# Harbor

[![Build Status](https://travis-ci.org/vmware/harbor.svg?branch=master)](https://travis-ci.org/vmware/harbor)
[![Coverage Status](https://coveralls.io/repos/github/vmware/harbor/badge.svg?branch=dev)](https://coveralls.io/github/vmware/harbor?branch=dev)

<img alt="Harbor" src="docs/img/harbor_logo.png">

Project Harbor is an enterprise-class registry server that stores and distributes Docker images. Harbor extends the open source Docker Distribution by adding the functionalities usually required by an enterprise, such as security, identity and management. As an enterprise private registry, Harbor offers better performance and security. Having a registry closer to the build and run environment improves the image transfer efficiency. Harbor supports the setup of multiple registries and has images replicated between them. With Harbor, the images are stored within the private registry, keeping the bits and intellectual properties behind the company firewall. In addition, Harbor offers advanced security features, such as user management, access control and activity auditing.

### Features
* **Role based access control**: Users and repositories are organized via 'projects' and a user can have different permission for images under a project.
* **Image replication**: Images can be replicated (synchronized) between multiple registry instances. Great for load balancing, high availability, hybrid and multi-cloud scenarios.
* **Graphical user portal**: User can easily browse, search repositories and manage projects.
* **AD/LDAP support**: Harbor integrates with existing enterprise AD/LDAP for user authentication and management.
* **Auditing**: All the operations to the repositories are tracked.
* **RESTful API**: RESTful APIs for most administrative operations, easy to integrate with external systems.
* **Easy deployment**: docker compose and offline installer.

### Install

**System requirements:**
Harbor only works with docker 1.10.0+ and docker-compose 1.6.0+.

#### Install via docker compose
On an Internet connected host, Harbor can be easily installed via docker-compose: 

1. Get the source code:
    
    ```sh
    $ git clone https://github.com/vmware/harbor
    ```
2. Edit the file **Deploy/harbor.cfg**, make necessary configuration changes such as hostname, admin password and mail server. Refer to [Installation and Configuration Guide](docs/installation_guide.md) for more info.  


3. Install Harbor with the following commands. Note that the docker-compose process can take a while.
    ```sh
    $ cd Deploy
    
    $ ./prepare
    Generated configuration file: ./config/ui/env
    Generated configuration file: ./config/ui/app.conf
    Generated configuration file: ./config/registry/config.yml
    Generated configuration file: ./config/db/env
    
    $ docker-compose up -d
    ```

#### Install via offline installer
For those who do not want to clone the source, or need to install Harbor on a server not connected to the Internet, there is a pre-built installation package available. For details on how to download and use the installation package, please refer to [Installation and Configuration Guide](docs/installation_guide.md).

#### After installation
_If everything worked properly, you should be able to open a browser to visit the admin portal at http://reg.yourdomain.com. Note that the default administrator username/password are admin/Harbor12345._

Log in to the admin portal and create a new project, e.g. `myproject`. You can then use docker commands to login and push images (by default, the registry server listens on port 80):
```sh
$ docker login reg.yourdomain.com
$ docker push reg.yourdomain.com/myproject/myrepo:mytag
```

### Upgrade

If you are upgrading Harbor from an older version with existing data, you need to migrate the data to fit the new database schema. For more details, please refer to [Data Migration Guide](docs/migration_guide.md).

### Run
For information on how to use Harbor, please take a look at [User Guide](docs/user_guide.md).

### Community
Get connected with Project Harbor's community and sign up with VMware {code} [https://code.vmware.com/join/](https://code.vmware.com/join/) to get invited to VMware {code} Slack group, Channel: #harbor. **Email:** harbor @ vmware.com . 

### Contribution
We welcome contributions from the community. If you wish to contribute code and you have not signed our contributor license agreement (CLA), our bot will update the issue when you open a pull request. For any questions about the CLA process, please refer to our [FAQ](https://cla.vmware.com/faq).

### License
Harbor is available under the [Apache 2 license](LICENSE).

This project uses open source components which have additional licensing terms.  The official docker images and licensing terms for these open source components can be found at the following locations:

* Photon OS 1.0: [docker image](https://hub.docker.com/_/photon/), [license](https://github.com/vmware/photon/blob/master/COPYING)
* Docker Registry 2.5: [docker image](https://hub.docker.com/_/registry/), [license](https://github.com/docker/distribution/blob/master/LICENSE)
* MySQL 5.6: [docker image](https://hub.docker.com/_/mysql/), [license](https://github.com/docker-library/mysql/blob/master/LICENSE)
* NGINX 1.9: [docker image](https://hub.docker.com/_/nginx/), [license](https://github.com/nginxinc/docker-nginx/blob/master/LICENSE)

### Partners
<a href="https://www.shurenyun.com/" border="0" target="_blank"><img alt="DataMan" src="docs/img/dataman.png"></a> &nbsp; &nbsp; <a href="http://www.slamtec.com" target="_blank" border="0"><img alt="SlamTec" src="docs/img/slamteclogo.png"></a>
&nbsp; &nbsp; <a href="https://www.caicloud.io" border="0"><img alt="CaiCloud" src="docs/img/caicloudLogoWeb.png"></a>

### Users
<a href="https://www.madailicai.com/" border="0" target="_blank"><img alt="MaDaiLiCai" src="docs/img/UserMaDai.jpg"></a> <a href="https://www.dianrong.com/" border="0" target="_blank"><img alt="Dianrong" src="docs/img/dianrong.png"></a>

### Supporting Technologies
<img alt="beego" src="docs/img/beegoLogo.png"> Harbor is powered by <a href="http://beego.me/">Beego</a>.

### About
Project Harbor is initiated by the Advanced Technology Center (ATC), VMware China R&D.
