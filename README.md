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

### Getting Started
Harbor is self-contained and can be easily deployed via docker-compose. The below are quick-start steps. Refer to the [Installation and Configuration Guide](docs/installation_guide.md) for detail information.  

**System requirements:**  
Harbor only works with docker 1.10+ and docker-compose 1.6.0+ .
The host must be connected to the Internet.

1. Get the source code:
    
    ```sh
    $ git clone https://github.com/vmware/harbor
    ```
2. Edit the file **Deploy/harbor.cfg**, make necessary configuration changes such as hostname, admin password and mail server. Refer to [Installation and Configuration Guide](docs/installation_guide.md) for more info.  


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

  If everything works fine, you can open a browser to visit the admin portal at http://reg.yourdomain.com . The default administrator username and password are admin/Harbor12345 .

  Create a new project, e.g. myproject, in the admin portal. You can then use docker commands to login and push images. The default port of Harbor registry server is 80:
  ```sh
  $ docker login reg.yourdomain.com
  $ docker push reg.yourdomain.com/myproject/myrepo
  ```

4. Deploy harbor on kubernetes.
  For now, it's a little tricky to start harbor on kubernetes because
     1. registry uses https, so we need cert or workaround to avoid errors like this:
        
        ```
        Error response from daemon: invalid registry endpoint https://{HOST}/v0/: unable to ping registry endpoint https://{HOST}/v0/
        v2 ping attempt failed with error: Get https://{HOST}/v2/: EOF
        v1 ping attempt failed with error: Get https://{HOST}/v1/_ping: EOF. If this private registry supports only HTTP or HTTPS with an unknown CA certificate, please add `--insecure-registry {HOST}` to the daemon's arguments. In the case of HTTPS, if you have access to the registry's CA certificate, no need for the flag; simply place the CA certificate at /etc/docker/certs.d/{HOST}/ca.crt
        ```
        
        There is a workaround if you don't have a cert. The workaround is to add the host into the list of insecure registry by editting the ```/etc/default/docker``` file:
        ```
        sudo vi /etc/default/docker
        ```
        add the line at the end of file:
        ```
        DOCKER_OPTS="$DOCKER_OPTS --insecure-registry={HOST}"
        ```
        and restart docker service
        ```
        sudo service docker restart
        ```
     2. The registry config file need to know the IP (or DNS name) of the registry, but on kubernetes, you won't know the IP before the service is created. There are several workarounds to solve this problem for now:
        - Use DNS name and link th DNS name with the IP after the service is created.
        - Rebuild the registry image with the service IP after the service is created and use ```kubectl rolling-update``` to update to the new image.
        
  To start harbor on kubernetes, you first need to change the host name at the registry config file and build the images by running:
  ```
  cd Deploy
  docker-compose build
  docker build -f kubernetes/dockerfiles/proxy-dockerfile -t {your_account}/proxy .
  docker build -f kubernetes/dockerfiles/registry-dockerfile -t {your_account}/registry .
  docker build -f kubernetes/dockerfiles/ui-dockerfile -t {your_account}/deploy_ui .
  docker tag deploy_mysql {your_account}/deploy_mysql
  docker push {your_account}/proxy
  docker push {your_account}/registry
  docker push {your_account}/deploy_ui
  docker push {your_account}/deploy_mysql
  ```
  
  where "your_account" is your own registry. Then you need to update the "image" field in the ```*-rc.yaml``` files at:
  ```
  Deploy/kubernetes/mysql-rc.yaml
  Deploy/kubernetes/proxy-rc.yaml
  Deploy/kubernetes/registry-rc.yaml
  Deploy/kubernetes/ui-rc.yaml
  ```

  Finally you can start the jobs by running:
  ```
  kubectl create -f Deploy/kubernetes
  ```

**NOTE:**  
To simplify the installation process, a pre-built installation package of Harbor is provided so that you don't need to clone the source code. By using this package, you can even install Harbor onto a host that is not connected to the Internet. For details on how to download and use this installation package, please refer to [Installation and Configuration Guide](docs/installation_guide.md) .

For information on how to use Harbor, please see [User Guide](docs/user_guide.md) .

### Contribution
We welcome contributions from the community. If you wish to contribute code and you have not signed our contributor license agreement (CLA), our bot will update the issue when you open a pull request. For any questions about the CLA process, please refer to our [FAQ](https://cla.vmware.com/faq).

### License
Harbor is available under the [Apache 2 license](LICENSE).

### Partners
<a href="https://www.shurenyun.com/" border="0" target="_blank"><img alt="DataMan" src="docs/img/dataman.png"></a> &nbsp; &nbsp; <a href="http://www.slamtec.com" target="_blank" border="0"><img alt="SlamTec" src="docs/img/slamteclogo.png"></a>

### Users
<a href="https://www.madailicai.com/" border="0" target="_blank"><img alt="MaDaiLiCai" src="docs/img/UserMaDai.jpg"></a>
