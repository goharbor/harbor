# Harbor

Harbor is a project to provide enterprise capabilities for Docker Registry V2.  It wraps the registry server to provide authorization and user interface.

### Features
* **Role Based Access Control**: Users and docker repositories are organized via "projects", a user can have differernt permission for images under a namespace.
* **Convenient user interface**: User can easily browse, search docker repositories, manage projects/namepaces.
* **LDAP support**: harbor can easily integrate to the existing ldap of entreprise.
* **Audting**: All the access to the repositories hosted on Harbor are immediately recorded and can be used for auditing purpose.

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

### License
Harbor is available under the [Apache 2 license](License.txt).

