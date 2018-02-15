# Project Harbor by VMware

[Harbor](http://vmware.github.io/harbor/) is an enterprise-class registry server that stores and distributes Docker images. Harbor extends the open source Docker Distribution by adding the functionalities usually required by an enterprise, such as security, identity and management. As an enterprise private registry, Harbor offers better performance and security. Having a registry closer to the build and run environment improves the image transfer efficiency. Harbor supports the setup of multiple registries and has images replicated between them. In addition, Harbor offers advanced security features, such as user management, access control and activity auditing.

## Introduction

This is an experimental monolithic chart that installs and configures VMWare Harbor and its dependencies. The initial implementation of this includes all of the components required to run Harbor. As upstream harbor becomes more cloud native we will be able to break apart the monolith and utitlize helm dependencies.

## Prerequisites

- Kubernetes 1.7+ with Beta APIs enabled
- PV provisioner support in the underlying infrastructure

## Installing the Chart

To install the chart with the release name `my-release`:

```bash
$ git clone https://github.com/vmware/harbor.git
$ cd harbor/contrib/helm/harbor
$ helm install --name my-release incubator/harbor
```

The command deploys Harbor on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```bash
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following tables lists the configurable parameters of the Percona chart and their default values.

| Parameter                  | Description                        | Default                 |
| -----------------------    | ---------------------------------- | ----------------------- |
| **Harbor** |
| `externalDomain`               | domain harbor will run on (https://*harbor.url*/) |`harbor.192.168.99.100.xip.io`   |
| `tls_crt`               | TLS certificate to use for Harbor's https endpoint | see values.yaml |
| `tls_key`               | TLS key to use for Harbor's https endpoint | see values.yaml |
| `ca_crt`               | CA Cert for self signed TLS cert | see values.yaml |
| `persistence.enabled` | enable persistent data storage | `false` |
| **Adminserver** |
| `adminserver.image.repository` | Repository for adminserver image | `vmware/harbor-adminserver` |
| `adminserver.image.tag` | Tag for adminserver image | `v1.3.0` |
| `adminserver.image.pullPolicy` | Pull Policy for adminserver image | `IfNotPresent` |
| `adminserver.emailHost` | email server | `smtp.mydomain.com` |
| `adminserver.emailPort` | email port | `25` |
| `adminserver.emailUser` | email username | `sample_admin@mydomain.com` |
| `adminserver.emailSsl` | email uses SSL? | `false` |
| `adminserver.emailFrom` | send email from address | `admin <sample_admin@mydomain.com>` |
| `adminserver.emailIdentity` | | "" |
| `adminserver.key` | adminsever key | `not-a-secure-key` |
| `adminserver.emailPwd` | password for email | `not-a-secure-password` |
| `adminserver.harborAdminPassword` | password for admin user | `Harbor12345` |
| `adminserver.resources` | [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) to allocate for container   | undefined |
| `adminserver.volumes` | used to create PVCs if persistence is enabled (see instructions in values.yaml) | see values.yaml |
| **Jobservice** |
| `jobservice.image.repository` | Repository for jobservice image | `vmware/harbor-jobservice` |
| `jobservice.image.tag` | Tag for jobservice image | `v1.3.0` |
| `jobservice.image.pullPolicy` | Pull Policy for jobservice image | `IfNotPresent` |
| `jobservice.key` | jobservice key | `not-a-secure-key` |
| `jobservice.secret` | jobservice secret | `not-a-secure-secret` |
| `jobservice.resources` | [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) to allocate for container   | undefined |
| **UI** |
| `ui.image.repository` | Repository for ui image | `vmware/harbor-ui` |
| `ui.image.tag` | Tag for ui image | `v1.3.0` |
| `ui.image.pullPolicy` | Pull Policy for ui image | `IfNotPresent` |
| `ui.key` | ui key | `not-a-secure-key` |
| `ui.secret` | ui secret | `not-a-secure-secret` |
| `ui.privateKeyPem` | ui private key | see values.yaml |
| `ui.resources` | [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) to allocate for container   | undefined |
| **MySQL** |
| `mysql.image.repository` | Repository for mysql image | `vmware/harbor-mysql` |
| `mysql.image.tag` | Tag for mysql image | `v1.3.0` |
| `mysql.image.pullPolicy` | Pull Policy for mysql image | `IfNotPresent` |
| `mysql.host` | MySQL Server | `~` |
| `mysql.port` | MySQL Port | `3306` |
| `mysql.user` | MySQL Username | `root` |
| `mysql.pass` | MySQL Password | `registry` |
| `mysql.database` | MySQL Database | `registry` |
| `mysql.resources` | [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) to allocate for container   | undefined |
| `mysql.volumes` | used to create PVCs if persistence is enabled (see instructions in values.yaml) | see values.yaml |
| **Registry** |
| `registry.image.repository` | Repository for registry image | `vmware/harbor-registry` |
| `registry.image.tag` | Tag for registry image | `v1.3.0` |
| `registry.image.pullPolicy` | Pull Policy for registry image | `IfNotPresent` |
| `registry.rootCrt` | registry root cert | see values.yaml |
| `registry.httpSecret` | registry secret | `not-a-secure-secret` |
| `registry.resources` | [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) to allocate for container   | undefined |
| `registry.volumes` | used to create PVCs if persistence is enabled (see instructions in values.yaml) | see values.yaml |
| **Clair** |
| `clair.enabled` | Enable clair? | `false` |
| `clair.postgresPassword` | password for clair postgres | see values.yaml |
| `clair.image.repository` | Repository for clair image | `vmware/clair` |
| `clair.image.tag` | Tag for clair image | `v2.0.1-photon` |
| `clair.image.pullPolicy` | Pull Policy for clair image | `IfNotPresent` |
| `clair.pgImage.repository` | Repository for clair postgres image | `postgres` |
| `clair.pgImage.tag` | Tag for clair postgres image | `9.6.4` |
| `clair.pgImage.pullPolicy` | Pull Policy for clair postgres image | `IfNotPresent` |
| `clair.resources` | [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) to allocate for container   | undefined | `clair.pgResources` | [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) to allocate for container   | undefined |
| | | |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example:

```bash
$ helm install --name my-release --set mysql.pass=baconeggs .
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
$ helm install --name my-release -f /path/to/values.yaml .
```

> **Tip**: You can use the default [values.yaml](values.yaml)

## Persistence

VMWare Harbor stores the data and configurations in emptyDir volumes. You can change the values.yaml to enable persistence and use a PersistentVolumeClaim instead.

> *"An emptyDir volume is first created when a Pod is assigned to a Node, and exists as long as that Pod is running on that node. When a Pod is removed from a node for any reason, the data in the emptyDir is deleted forever."*
