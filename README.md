# Harbor

[![Build Status](https://travis-ci.org/vmware/harbor.svg?branch=master)](https://travis-ci.org/vmware/harbor)
[![Coverage Status](https://coveralls.io/repos/github/vmware/harbor/badge.svg?branch=dev)](https://coveralls.io/github/vmware/harbor?branch=dev)

<img alt="Harbor" src="docs/img/harbor_logo.png">

Project Harbor is an enterprise-class registry server that stores and distributes Docker images. Harbor extends the open source Docker Distribution by adding the functionalities usually required by an enterprise, such as security, identity and management. As an enterprise private registry, Harbor offers better performance and security. Having a registry closer to the build and run environment improves the image transfer efficiency. Harbor supports the setup of multiple registries and has images replicated between them. In addition, Harbor offers advanced security features, such as user management, access control and activity auditing.

### Features
* **Role based access control**: Users and repositories are organized via 'projects' and a user can have different permission for images under a project.
* **Policy based image replication**: Images can be replicated (synchronized) between multiple registry instances. Great for load balancing, high availability, multi-datacenter, hybrid and multi-cloud scenarios.
* **LDAP/AD support**: Harbor integrates with existing enterprise LDAP/AD for user authentication and management.
* **Image deletion & garbage collection**: Images can be deleted and their space can be recycled. 
* **Notary**: Image authenticity can be ensured. 
* **Graphical user portal**: User can easily browse, search repositories and manage projects.
* **Auditing**: All the operations to the repositories are tracked.
* **RESTful API**: RESTful APIs for most administrative operations, easy to integrate with external systems.
* **Easy deployment**: Provide both an online and offline installer. 

### Install & Run

**System requirements:**

**On a Linux host:** docker 1.10.0+ and docker-compose 1.6.0+ .

Download binaries of **[Harbor release ](https://github.com/vmware/harbor/releases)** and follow **[Installation & Configuration Guide](docs/installation_guide.md)** to install Harbor.

Refer to **[User Guide](docs/user_guide.md)** for more details on how to use Harbor.

### Community
**Slack:** Join Harbor's community here: [VMware {code}](https://code.vmware.com/join/), Channel: #harbor.  
**Email:** harbor@ vmware.com .  
**WeChat Group:** Add WeChat id *connect1688* to join WeChat discussion group.  
More info on [partners and users](partners.md). 

### Contribution
We welcome contributions from the community. If you wish to contribute code and you have not signed our contributor license agreement (CLA), our bot will update the issue when you open a pull request. For any questions about the CLA process, please refer to our [FAQ](https://cla.vmware.com/faq). Contact us for any questions: harbor @vmware.com .

### License
Harbor is available under the [Apache 2 license](LICENSE).

This project uses open source components which have additional licensing terms.  The official docker images and licensing terms for these open source components can be found at the following locations:

* Photon OS 1.0: [docker image](https://hub.docker.com/_/photon/), [license](https://github.com/vmware/photon/blob/master/COPYING)
* Docker Registry 2.6: [docker image](https://hub.docker.com/_/registry/), [license](https://github.com/docker/distribution/blob/master/LICENSE)
* MySQL 5.6: [docker image](https://hub.docker.com/_/mysql/), [license](https://github.com/docker-library/mysql/blob/master/LICENSE)
* NGINX 1.11.5: [docker image](https://hub.docker.com/_/nginx/), [license](https://github.com/nginxinc/docker-nginx/blob/master/LICENSE)
