# Harbor

[![Build Status](https://travis-ci.org/vmware/harbor.svg?branch=master)](https://travis-ci.org/vmware/harbor)
[![Coverage Status](https://coveralls.io/repos/github/vmware/harbor/badge.svg?branch=master)](https://coveralls.io/github/vmware/harbor?branch=master)

**Note**: The `master` branch may be in an *unstable or even broken state* during development.
Please use [releases](https://github.com/vmware/harbor/releases) instead of the `master` branch in order to get stable binaries.

<img alt="Harbor" src="docs/img/harbor_logo.png">

Project Harbor is an an open source trusted cloud native registry project that stores, signs, and scans content. Harbor extends the open source Docker Distribution by adding the functionalities usually required by users such as security, identity and management. Having a registry closer to the build and run environment can improve the image transfer efficiency. Harbor supports replication of images between registries, and also offers advanced security features such as user management, access control and activity auditing.

### Features
* **Role based access control**: Users and repositories are organized via 'projects' and a user can have different permission for images under a project.
* **Policy based image replication**: Images can be replicated (synchronized) between multiple registry instances, with auto-retry on errors. Great for load balancing, high availability, multi-datacenter, hybrid and multi-cloud scenarios.
* **Vulnerability Scanning**: Harbor scans images regularly and warns users of vulnerabilities.
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
**Twitter:** [@project_harbor](https://twitter.com/project_harbor)
**User Group:** Join Harbor user email group: [harbor-users@googlegroups.com](https://groups.google.com/forum/#!forum/harbor-users) to get update of Harbor's news, features, releases, or to provide suggestion and feedback. To subscribe, send an email to [harbor-users+subscribe@googlegroups.com](mailto:harbor-users+subscribe@googlegroups.com) .  
**Developer Group:** Join Harbor developer group: [harbor-dev@googlegroups.com](https://groups.google.com/forum/#!forum/harbor-dev) for discussion on Harbor development and contribution. To subscribe, send an email to [harbor-dev+subscribe@googlegroups.com](mailto:harbor-dev+subscribe@googlegroups.com).  
**Slack:** Join Harbor's community for discussion and ask questions: [VMware {code}](https://code.vmware.com/join/), channel: [#harbor](https://vmwarecode.slack.com/messages/harbor).

More info on [partners and users](partners.md).

### Contribution
We welcome contributions from the community. If you wish to contribute code and you have not signed our contributor license agreement (CLA), our bot will update the issue when you open a pull request. For any questions about the CLA process, please refer to our [FAQ](https://cla.vmware.com/faq). Contact us for any questions: [harbor@vmware.com](mailto:harbor@vmware.com).

### Demos
* ![play](docs/img/video.png) **Content Trust** ( [youtube](https://www.youtube.com/watch?v=pPklSTJZY2E) , [Tencent Video](https://v.qq.com/x/page/n0553fzzrnf.html) )
* ![play](docs/img/video.png) **Role Based Access Control** ( [youtube](https://www.youtube.com/watch?v=2ZIu9XTvsC0) , [Tencent Video](https://v.qq.com/x/page/l0553yw19ek.html) )
* ![play](docs/img/video.png) **Vulnerability Scanning** ( [youtube](https://www.youtube.com/watch?v=K4tJ6B2cGR4) , [Tencent Video](https://v.qq.com/x/page/s0553k9692d.html) )
* ![play](docs/img/video.png) **Image Replication** ( [youtube](https://www.youtube.com/watch?v=1NPlzrm5ozE) , [Tencent Video](https://v.qq.com/x/page/a0553wc7fs9.html) )
* ![play](docs/img/video.png) **VMworld 2017** ( [youtube](https://www.youtube.com/watch?v=tI5xMe24fJ4) )

### License
Harbor is available under the [Apache 2 license](LICENSE).

This project uses open source components which have additional licensing terms.  The official docker images and licensing terms for these open source components can be found at the following locations:

* Photon OS 1.0: [docker image](https://hub.docker.com/_/photon/), [license](https://github.com/vmware/photon/blob/master/COPYING)
* MySQL 5.6: [docker image](https://hub.docker.com/_/mysql/), [license](https://github.com/docker-library/mysql/blob/master/LICENSE)

### Commercial Support
If you need commercial support of Harbor, please contact us for more information: [harbor@vmware.com](mailto:harbor@vmware.com).
