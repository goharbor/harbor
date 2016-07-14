# Harbor

[![Build Status](https://travis-ci.org/vmware/harbor.svg?branch=master)](https://travis-ci.org/vmware/harbor)

![alg tag](https://cloud.githubusercontent.com/assets/2390463/13484557/088a1000-e13a-11e5-87d4-a64366365bef.png)

> Project Harbor is initiated by VMware China R&D as a Cloud Application Accelerator (CAA) project. CAA provides a set of tools to improve the productivity of cloud developers in China and other countries. CAA includes tools like registry server, mirror server, decentralized image distributor, etc.

Project Harbor is an enterprise-class registry server, which extends the open source Docker Registry server by adding the functionality usually required by an enterprise, such as security, control, and management. Harbor is primarily designed to be a private registry - providing the needed security and control that enterprises require.  It also helps minimize bandwidth usage, which is helpful to both improve productivity (local network access) as well as performance (for those with poor internet connectivity).

### Features
* **Role based access control**: Users and Docker repositories are organized via "projects", a user can have different permission for images under a project.
* **Image replication**: Images can be replicated(synchronized) between multiple registry instances. Great for load balancing and distributed data centers.
* **Graphical user portal**: User can easily browse, search Docker repositories, manage projects/namespaces.
* **AD/LDAP support**: Harbor integrates with existing enterprise AD/LDAP for user authentication and management.
* **Auditing**: All the operations to the repositories are tracked.
* **Internationalization**: Already localized for English, Chinese, German, Japanese and Russian. More languages can be added.
* **RESTful API**: RESTful APIs for most administrative operations, easing intergration with external management platforms.

### Getting Started
Harbor is self-contained and can be easily deployed via docker-compose (Quick-Start steps below). Refer to the [Installation and Configuration Guide](docs/installation_guide.md) for detailed information.  

**System requirements:**  
Harbor only works with docker 1.10+ and docker-compose 1.6.0+, and an internet-connected host.

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

_If everything worked properly, you should be able to open a browser to visit the admin portal at http://reg.yourdomain.com . Note that the default administrator username/password are admin/Harbor12345 ._

Log in to the admin portal and create a new project, e.g. `myproject`. You can then use docker commands to login and push images (By default, the registry server listens on port 80):
```sh
$ docker login reg.yourdomain.com
$ docker push reg.yourdomain.com/myproject/myrepo
```

**Data migration:**

If you are upgrading Harbor from an older version with existing data, you may need to migrate the data to fit the new database schema. For more details please refer to [Data Migration Guide](docs/migration_guide.md) .

**NOTE:**  
For those who don't want to clone the source, or need to install Harbor on a server not connected to the Internet - there is a pre-built installation package available. For details on how to download and use this installation package, please refer to [Installation and Configuration Guide](docs/installation_guide.md) .

For information on how to use Harbor, please see [User Guide](docs/user_guide.md) .

### Contribution
We welcome contributions from the community. If you wish to contribute code and you have not signed our contributor license agreement (CLA), our bot will update the issue when you open a pull request. For any questions about the CLA process, please refer to our [FAQ](https://cla.vmware.com/faq).

### License
Harbor is available under the [Apache 2 license](LICENSE).

### Partners
<a href="https://www.shurenyun.com/" border="0" target="_blank"><img alt="DataMan" src="docs/img/dataman.png"></a> &nbsp; &nbsp; <a href="http://www.slamtec.com" target="_blank" border="0"><img alt="SlamTec" src="docs/img/slamteclogo.png"></a>
&nbsp; &nbsp; <a href="https://www.caicloud.io" border="0"><img alt="CaiCloud" src="docs/img/caicloudLogoWeb.png"></a>

### Users
<a href="https://www.madailicai.com/" border="0" target="_blank"><img alt="MaDaiLiCai" src="docs/img/UserMaDai.jpg"></a> <a href="https://www.dianrong.com/" border="0" target="_blank"><img alt="Dianrong" src="docs/img/dianrong.png"></a>

### Supporting Technologies
<img alt="beego" src="docs/img/beegoLogo.png"> Harbor is powered by <a href="http://beego.me/">Beego</a>, an open source framework to build and develop applications in the Go way.
