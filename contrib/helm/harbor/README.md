# Helm Chart for Harbor

## Introduction

This [Helm](https://github.com/kubernetes/helm) chart installs [Harbor](http://vmware.github.io/harbor/) in a Kubernetes cluster. Currently this chart supports Harbor v1.4.0 release. Welcome to [contribute](CONTRIBUTING.md) to Helm Chart for Harbor.

## Prerequisites

- Kubernetes cluster 1.8+ with Beta APIs enabled
- Kubernetes Ingress Controller is enabled
- kubectl CLI 1.8+
- Helm CLI 2.8.0+

## Known Issues

- This chart doesn't work with Kubernetes security update release 1.8.9+ and 1.9.4+. Refer to [issue 4496](https://github.com/vmware/harbor/issues/4496).

## Setup a Kubernetes cluster

You can use any tools to setup a K8s cluster.
In this guide, we use [minikube](https://github.com/kubernetes/minikube) 0.25.0 to setup a K8s cluster as the dev/test env.
```bash
# Start minikube
minikube start --vm-driver=none
# Enable Ingress Controller
minikube addons enable ingress
```
## Installing the Chart

First install [Helm CLI](https://github.com/kubernetes/helm#install), then initialize Helm.
```bash
helm init
```
Download Harbor helm chart code.
```bash
git clone https://github.com/vmware/harbor
cd contrib/helm/harbor
```
Download external dependent charts required by Harbor chart.
```bash
helm dependency update
```
### Secure Registry Mode

By default this chart will generate a root CA and SSL certificate for your Harbor.
You can also use your own CA signed certificate:

open values.yaml, set the value of 'externalDomain' to your Harbor FQDN, and
set value of 'tlsCrt', 'tlsKey', 'caCrt'. The common name of the certificate must match your Harbor FQDN.

Install the Harbor helm chart with a release name `my-release`:
```bash
helm install . --debug --name my-release --set externalDomain=harbor.my.domain
```
**Make sure** `harbor.my.domain` resolves to the K8s Ingress Controller IP on the machines where you run docker or access Harbor UI.
You can add `harbor.my.domain` and IP mapping in the DNS server, or in /etc/hosts, or use the FQDN `harbor.<IP>.xip.io`.

Follow the `NOTES` section in the command output to get Harbor admin password and **add Harbor root CA into docker trusted certificates**.

The command deploys Harbor on the Kubernetes cluster in the default configuration.
The [configuration](#configuration) section lists the parameters that can be configured in values.yaml or via '--set' params during installation.

> **Tip**: List all releases using `helm list`

### Insecure Registry Mode

If setting Harbor Registry as insecure-registries for docker,
you don't need to generate Root CA and SSL certificate for the Harbor ingress controller.

Install the Harbor helm chart with a release name `my-release`:
```bash
helm install . --debug --name my-release --set externalDomain=harbor.my.domain,insecureRegistry=true
```
**Make sure** `harbor.my.domain` resolves to the K8s Ingress Controller IP on the machines where you run docker or access Harbor UI.
You can add `harbor.my.domain` and IP mapping in the DNS server, or in /etc/hosts, or use the FQDN `harbor.<IP>.xip.io`.

Then add `"insecure-registries": ["harbor.my.domain"]` in the docker daemon config file and restart docker service.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```bash
helm delete my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following tables lists the configurable parameters of the Harbor chart and the default values.

| Parameter                  | Description                        | Default                 |
| -----------------------    | ---------------------------------- | ----------------------- |
| **Harbor** |
| `harborImageTag`     | The tag for Harbor docker images | `v1.4.0` |
| `externalDomain`       | Harbor will run on (https://`externalDomain`/). Recommend using K8s Ingress Controller FQDN as `externalDomain`, or make sure this FQDN resolves to the K8s Ingress Controller IP. | `harbor.my.domain` |
| `insecureRegistry`     | If set to true, you don't need to set tlsCrt/tlsKey/caCrt, but must add Harbor FQDN as insecure-registries for your docker client. | `false` |
| `tlsCrt`               | TLS certificate to use for Harbor's https endpoint. Its CN must match `externalDomain`. | auto-generated |
| `tlsKey`               | TLS key to use for Harbor's https endpoint | auto-generated |
| `caCrt`                | CA Cert for self signed TLS cert | auto-generated |
| `persistence.enabled` | enable persistent data storage | `false` |
| `secretKey` | The secret key used for encryption. Must be a string of 16 chars. | `not-a-secure-key` |
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
| `adminserver.adminPassword` | password for admin user | `Harbor12345` |
| `adminserver.authenticationMode` | authentication mode for Harbor ( `db_auth` for local database, `ldap_auth` for LDAP, etc...) [Docs](https://github.com/vmware/harbor/blob/master/docs/user_guide.md#user-account) | `db_auth` |
| `adminserver.selfRegistration` | Allows users to register by themselves, otherwise only administrators can add users | `on` |
| `adminserver.ldap.url` | LDAP server URL for `ldap_auth` authentication | `ldaps://ldapserver` |
| `adminserver.ldap.searchDN` | LDAP Search DN | `` |
| `adminserver.ldap.baseDN` | LDAP Base DN | `` |
| `adminserver.ldap.filter` | LDAP Filter | `(objectClass=person)` |
| `adminserver.ldap.uid` | LDAP UID | `uid` |
| `adminserver.ldap.scope` | LDAP Scope | `2` |
| `adminserver.ldap.timeout` | LDAP Timeout | `5` |
| `adminserver.ldap.verifyCert` | LDAP Verify HTTPS Certificate | `True` |
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
| `registry.image.pullPolicy` | Pull Policy for registry image | `IfNotPresent` |
| `registry.rootCrt` | registry root cert | see values.yaml |
| `registry.httpSecret` | registry secret | `not-a-secure-secret` |
| `registry.resources` | [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) to allocate for container   | undefined |
| `registry.volumes` | used to create PVCs if persistence is enabled (see instructions in values.yaml) | see values.yaml |
| **Clair** |
| `clair.enabled` | Enable clair? | `true` |
| `clair.image.repository` | Repository for clair image | `vmware/clair-photon` |
| `clair.image.tag` | Tag for clair image | `v2.0.1-v1.4.0`
| `clair.resources` | [resources](https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/) to allocate for container   | undefined
| `postgresql` | Overrides for postgresql chart [values.yaml](https://github.com/kubernetes/charts/blob/f2938a46e3ae8e2512ede1142465004094c3c333/stable/postgresql/values.yaml) | see values.yaml
| | | |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`. For example:

```bash
helm install --name my-release --set mysql.pass=baconeggs .
```

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
helm install --name my-release -f /path/to/values.yaml .
```

> **Tip**: You can use the default [values.yaml](values.yaml)

## Persistence

Harbor stores the data and configurations in emptyDir volumes. You can change the values.yaml to enable persistence and use a PersistentVolumeClaim instead.

> *"An emptyDir volume is first created when a Pod is assigned to a Node, and exists as long as that Pod is running on that node. When a Pod is removed from a node for any reason, the data in the emptyDir is deleted forever."*
