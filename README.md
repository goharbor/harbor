# Harbor

[![Build Status](https://travis-ci.org/vmware/harbor.svg?branch=master)](https://travis-ci.org/vmware/harbor)

![alg tag](https://cloud.githubusercontent.com/assets/2390463/13484557/088a1000-e13a-11e5-87d4-a64366365bef.png)

> Project Harbor is initiated by VMware China R&D as a Cloud Application Accelerator (CAA) project. CAA provides a set of tools to improve the productivity of cloud developers in China and other countries. CAA includes tools like registry server, mirror server, decentralized image distributor, etc.

Project Harbor is an enterprise-class registry server. It extends the open source Docker Registry server by adding more functionalities usually required by an enterprise. Harbor is designed to be deployed in a private environment of an organization. A private registry is important for organizations who care much about security. In addition, a private registry improves productivity by eliminating the need to download images from the public network. This is very helpful to container users who do not have a good network to the Internet. 

### Features
* **Role Based Access Control**: Users and docker repositories are organized via "projects", a user can have different permission for images under a namespace.
* **Graphical user portal**: User can easily browse, search docker repositories, manage projects/namespaces.
* **AD/LDAP support**: Harbor integrates with existing AD/LDAP of the enterprise for user authentication and management.
* **Auditing**: All the operations to the repositories are tracked and can be used for auditing purpose.
* **Internationalization**: Localized for English and Chinese languages. More languages can be added.
* **RESTful API**: RESTful APIs are provided for most administrative operations of Harbor. The integration with other management softwares becomes easy.

### Try it
Harbor is self-contained and can be easily deployed via docker-compose.  

**System requirements:**  
Harbor only works with docker 1.8+ and docker-compose 1.6.0+ .
The host must be connected to the Internet.

1. Get the source code:
    
    ```sh
    $ git clone https://github.com/vmware/harbor
    ```
2. Edit the file **Deploy/harbor.cfg**, make necessary configuration changes such as hostname, admin password and mail server. Refer to [Installation Guide](docs/installation_guide.md) for more info.  


3. Install Harbor by the following commands. It may take a while for the docker-compose process to finish.
    ```sh
    $ cd Deploy
    
    $ ./prepare
    Generated configuration file: ./config/ui/env
    Generated configuration file: ./config/ui/app.conf
    Generated configuration file: ./config/registry/config.yml
    Generated configuration file: ./config/db/env
    
    $ docker-compose up
    ```
If everything works fine, you can open a browser to visit the admin portal at http://your_registry_host . The default administrator username and password are admin/Harbor12345 .

After creating a project in the admin portal, you can login and use docker commands to push images.  The default port of Harbor registry server is 80:
```sh
$ docker login your_registry_host
$ docker push your_registry_host/myrepo/myapp
```

**NOTE:**  
To simplify the installation process, a pre-built installation package of Harbor is provided so that you don't need to clone the source code. By using this package, you can even install Harbor onto a host that is not connected to the Internet. For details on how to download and use this installation package, please refer to [Installation Guide](docs/installation_guide.md) .

For information on how to use Harbor, please see [User Guide](docs/user_guide.md) .

### Contribution
We welcome contributions from the community.  If you wish to contribute code, we require that you first sign our [Contributor License Agreement](https://vmware.github.io/photon/assets/files/vmware_cla.pdf) and return a copy to osscontributions@vmware.com before we can merge your contribution.

### License
Harbor is available under the [Apache 2 license](LICENSE).

### Partners
<a href="https://www.shurenyun.com/" border="0" target="_blank"><img alt="DataMan" src="docs/img/dataman.png"></a>

### Users
<a href="https://www.madailicai.com/" border="0" target="_blank"><img alt="MaDaiLiCai" src="docs/img/UserMaDai.jpg"></a>  <a href="http://www.slamtec.com" target="_blank" border="0"><img alt="SlamTec" src="docs/img/slamteclogo.png"></a>
