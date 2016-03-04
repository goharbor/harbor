# Harbor

![alg tag](https://cloud.githubusercontent.com/assets/2390463/13484557/088a1000-e13a-11e5-87d4-a64366365bef.png)

Project Harbor is an enterprise-class registry server. It extends the open source Docker Registry server by adding more functionalities usually required by an enterprise. Harbor is designed to be deployed in a private environment of an organization. A private registry is important for organizations who care much about security. In addition, a private registry improves productivity by eliminating the need to download images from public network. This is very helpful to container users who do not have a good network to the Internet. In particular, Harbor accelerates the progress of Chinese developers, because they no longer need to pull images from the Internet.

### Features
* **Role Based Access Control**: Users and docker repositories are organized via "projects", a user can have differernt permission for images under a namespace.
* **Graphical user portal**: User can easily browse, search docker repositories, manage projects/namepaces.
* **AD/LDAP support**: Harbor integrates with existing AD/LDAP of enterprise for user authentication and management.
* **Audting**: All the operations to the repositories are tracked and can be used for auditing purpose.
* **Internationalization**: Localized for English and Chinese languages. More languages can be added.
* **RESTful API**: RESTful APIs are provided for most administrative operations of Harbor. The integration with other management software becomes easy.

### Try it
Harbor is self contained and can be easily deployed via docker-compose.
```sh
$ cd Deploy
#make update to the parameters in ./harbor.cfg
$ ./prepare
Generated configuration file: ./config/ui/env
Generated configuration file: ./config/ui/app.conf
Generated configuration file: ./config/registry/config.yml
Generated configuration file: ./config/db/env
$ docker-compose up
```

### Contribution
We welcome contributions from the community.  If you wish to contribute code, we require that you first sign our [Contributor License Agreement](https://vmware.github.io/photon/assets/files/vmware_cla.pdf) and return a copy to osscontributions@vmware.com before we can merge your contribution.

### License
Harbor is available under the [Apache 2 license](LICENSE).

