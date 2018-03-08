# VMWare Harbor

[VMWare Harbor](http://vmware.github.io/harbor/) is an enterprise-class registry server that stores and distributes Docker images. Harbor extends the open source Docker Distribution by adding the functionalities usually required by an enterprise, such as security, identity and management. As an enterprise private registry, Harbor offers better performance and security. Having a registry closer to the build and run environment improves the image transfer efficiency. Harbor supports the setup of multiple registries and has images replicated between them. In addition, Harbor offers advanced security features, such as user management, access control and activity auditing.

## Introduction

This chart installs and configures VMWare Harbor.

## Prerequisites

- Kubernetes 1.7+ with Beta APIs enabled
- Kubernetes Ingress Controller is enabled
- PV provisioner support in the underlying infrastructure

## Setup a Kubernetes cluster

You can use any tools to setup a K8s cluster.
In this guide, we use [minikube](https://github.com/kubernetes/minikube) to setup a K8s cluster as the dev/test env.

```bash
# Start minikube
$ minikube start
# Enable Ingress Controller
$ minikube addons enable ingress
```

## Installing the Chart

First install [Helm CLI](https://github.com/kubernetes/helm#install), then initialize Helm.

```bash
$ helm init
```

Download Harbor helm chart code.

```bash
$ git clone https://github.com/vmware/harbor
$ cd harbor/contrib/helm/harbor
```

### Insecure Registry Mode
_Not recommended._

If setting Harbor Registry as insecure-registries for docker,
you don't need to generate Root CA and SSL certificate for the Harbor ingress controller.

Install the Harbor helm chart with a release name `my-release`:

```bash
$ helm install . --debug --name my-release \
  --set externalDomain=harbor.my.domain,insecureRegistry=true
```

**Make sure** `harbor.my.domain` resolves to the K8s Ingress Controller IP on the machines where you run docker or access Harbor UI.
You can add `harbor.my.domain` and IP mapping in the DNS server, or in /etc/hosts, or use the FQDN `harbor.<IP>.xip.io`.

Then add `"insecure-registries": ["harbor.my.domain"]` in the docker daemon config file and restart docker service.

### Secure Registry Mode

If you are deploying in minikube and your minikube IP is `192.168.99.100` you can use the default Certificates and skip the sections.

Generate Root CA and SSL certificate for your Harbor.
You can use your own certificate or follow this [guide](https://datacenteroverlords.com/2012/03/01/creating-your-own-ssl-certificate-authority/)
to create a self-signed certificate. The common name of the certificate must match your Harbor FQDN.

Open values.yaml, set the value of 'externalDomain' to your Harbor FQDN, and
set value of 'tlsCrt', 'tlsKey', 'caCrt' to the generated certificate.

Install the Harbor helm chart with a release name `my-release`:

```bash
$ helm install . --debug --name my-release
```

Follow the `NOTES` section in the command output to get Harbor admin password and **add Harbor root CA into docker trusted certificates**.

The command deploys Harbor on the Kubernetes cluster in the default configuration.
The [configuration](#configuration) section lists the parameters that can be configured during installation.

> **Tip**: List all releases using `helm list`

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```bash
$ helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following tables lists the configurable parameters of the Harbor chart and the default values.

| Parameter                  | Description                        | Default                 |
| -----------------------    | ---------------------------------- | ----------------------- |
| **Harbor** |
| `externalDomain`       | Harbor will run on (https://*externalDomain*/). Make sure this FQDN resolves to the K8s Ingress Controller IP. | undefined |
| `insecureRegistry`     | Set to true if setting Harbor Registry as insecure-registries for docker | `false` |
| `tlsCrt`               | TLS certificate to use for Harbor's https endpoint | see values.yaml |
| `tlsKey`               | TLS key to use for Harbor's https endpoint | see values.yaml |
| `caCrt`                | CA Cert for self signed TLS cert | see values.yaml |
| `ingress.annotations`  | Annotations for ingress controller | see values.yaml |
| `persistence.enabled` | enable persistent data storage | `false` |
| **Adminserver** |
| `adminserver.image.repository` | Repository for adminserver image | `vmware/harbor-adminserver` |
| `adminserver.image.tag` | Tag for adminserver image | `v1.4.0` |
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
| `jobservice.image.tag` | Tag for jobservice image | `v1.4.0` |
| `jobservice.image.pullPolicy` | Pull Policy for jobservice image | `IfNotPresent` |
| `jobservice.key` | jobservice key | `not-a-secure-key` |
| `jobservice.secret` | jobservice secret | `not-a-secure-secret` |
| `jobservice.resources` | [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) to allocate for container   | undefined |
| **UI** |
| `ui.image.repository` | Repository for ui image | `vmware/harbor-ui` |
| `ui.image.tag` | Tag for ui image | `v1.4.0` |
| `ui.image.pullPolicy` | Pull Policy for ui image | `IfNotPresent` |
| `ui.key` | ui key | `not-a-secure-key` |
| `ui.secret` | ui secret | `not-a-secure-secret` |
| `ui.privateKeyPem` | ui private key | see values.yaml |
| `ui.resources` | [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) to allocate for container   | undefined |
| **MySQL** |
| `mysql.image.repository` | Repository for mysql image | `vmware/harbor-mysql` |
| `mysql.image.tag` | Tag for mysql image | `v1.4.0` |
| `mysql.image.pullPolicy` | Pull Policy for mysql image | `IfNotPresent` |
| `mysql.host` | MySQL Server | `~` |
| `mysql.port` | MySQL Port | `3306` |
| `mysql.user` | MySQL Username | `root` |
| `mysql.pass` | MySQL Password | `registry` |
| `mysql.database` | MySQL Database | `registry` |
| `mysql.resources` | [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) to allocate for container   | undefined |
| `mysql.volumes` | used to create PVCs if persistence is enabled (see instructions in values.yaml) | see values.yaml |
| **Registry** |
| `registry.image.repository` | Repository for registry image | `vmware/registry-photon` |
| `registry.image.tag` | Tag for registry image | `v2.6.2-v1.4.0` |
| `registry.image.pullPolicy` | Pull Policy for admregistryinserver image | `IfNotPresent` |
| `registry.rootCrt` | registry root cert | see values.yaml |
| `registry.httpSecret` | registry secret | `not-a-secure-secret` |
| `registry.resources` | [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) to allocate for container   | undefined |
| `registry.volumes` | used to create PVCs if persistence is enabled (see instructions in values.yaml) | see values.yaml |
| **Clair** |
| `clair.enabled` | Enable clair? | `true` |
| `clair.image.repository` | Repository for clair image | `vmware/clair-photon` |
| `clair.image.tag` | Tag for clair image | `v2.0.1-v1.4.0` |
| `clair.image.pullPolicy` | pull policy for clair image | IfNotPresent |
| `clair.pgImage.repository` | Repository for clair postgres image | `vmware/clair-photon` |
| `clair.pgImage.tag` | Tag for clair postgres image | `v2.0.1-v1.4.0` |
| `clair.pgImage.pullPolicy` | pull policy for clair postgres image | IfNotPresent |
| `clair.postgresPassword` | password for clair postgres | see values.yaml |
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