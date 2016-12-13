# Clair

[![Build Status](https://api.travis-ci.org/coreos/clair.svg?branch=master "Build Status")](https://travis-ci.org/coreos/clair)
[![Docker Repository on Quay](https://quay.io/repository/coreos/clair/status "Docker Repository on Quay")](https://quay.io/repository/coreos/clair)
[![Go Report Card](https://goreportcard.com/badge/coreos/clair "Go Report Card")](https://goreportcard.com/report/coreos/clair)
[![GoDoc](https://godoc.org/github.com/coreos/clair?status.svg "GoDoc")](https://godoc.org/github.com/coreos/clair)
[![IRC Channel](https://img.shields.io/badge/freenode-%23clair-blue.svg "IRC Channel")](http://webchat.freenode.net/?channels=clair)

**Note**: The `master` branch may be in an *unstable or even broken state* during development.
Please use [releases] instead of the `master` branch in order to get stable binaries.

![Clair Logo](img/Clair_horizontal_color.png)

Clair is an open source project for the static analysis of vulnerabilities in [appc] and [docker] containers.

Vulnerability data is continuously imported from a known set of sources and correlated with the indexed contents of container images in order to produce lists of vulnerabilities that threaten a container.
When vulnerability data changes upstream, the previous state and new state of the vulnerability along with the images they affect can be sent via webhook to a configured endpoint.
All major components can be [customized programmatically] at compile-time without forking the project.

Our goal is to enable a more transparent view of the security of container-based infrastructure.
Thus, the project was named `Clair` after the French term which translates to *clear*, *bright*, *transparent*.

[appc]: https://github.com/appc/spec
[docker]: https://github.com/docker/docker/blob/master/image/spec/v1.md
[customized programmatically]: #customization
[releases]: https://github.com/coreos/clair/releases

## Common Use Cases

### Manual Auditing

You're building an application and want to depend on a third-party container image that you found by searching the internet.
To make sure that you do not knowingly introduce a new vulnerability into your production service, you decide to scan the container for vulnerabilities.
You `docker pull` the container to your development machine and start an instance of Clair.
Once it finishes updating, you use the [local image analysis tool] to analyze the container.
You realize this container is vulnerable to many critical CVEs, so you decide to use another one.

[local image analysis tool]: https://github.com/coreos/clair/tree/master/contrib/analyze-local-images

### Container Registry Integration

Your company has a continuous-integration pipeline and you want to stop deployments if they introduce a dangerous vulnerability.
A developer merges some code into the master branch of your codebase.
The first step of your continuous-integration pipeline automates the testing and building of your container and pushes a new container to your container registry.
Your container registry notifies Clair which causes the download and indexing of the images for the new container.
Clair detects some vulnerabilities and sends a webhook to your continuous deployment tool to prevent this vulnerable build from seeing the light of day.

## Hello Heartbleed

During the first run, Clair will bootstrap its database with vulnerability data from its data sources.
It can take several minutes before the database has been fully populated.

**NOTE:** These setups are not meant for production workloads, but as a quick way to get started.

### Kubernetes

An easy way to run Clair is with Kubernetes 1.2+.
If you are using the [CoreOS Kubernetes single-node instructions][single-node] for Vagrant you will be able to access the Clair's API at http://172.17.4.99:30060/ after following these instructions.

```
git clone https://github.com/coreos/clair
cd clair/contrib/k8s
kubectl create secret generic clairsecret --from-file=./config.yaml
kubectl create -f clair-kubernetes.yaml
```

[single-node]: https://coreos.com/kubernetes/docs/latest/kubernetes-on-vagrant-single.html

### Docker Compose

Another easy way to get an instance of Clair running is to use Docker Compose to run everything locally.
This runs a PostgreSQL database insecurely and locally in a container.
This method should only be used for testing.

```sh
$ curl -L https://raw.githubusercontent.com/coreos/clair/v1.2.5/docker-compose.yml -o $HOME/docker-compose.yml
$ mkdir $HOME/clair_config
$ curl -L https://raw.githubusercontent.com/coreos/clair/v1.2.5/config.example.yaml -o $HOME/clair_config/config.yaml
$ $EDITOR $HOME/clair_config/config.yaml # Edit database source to be postgresql://postgres:password@postgres:5432?sslmode=disable
$ docker-compose -f $HOME/docker-compose.yml up -d
```

Docker Compose may start Clair before Postgres which will raise an error.
If this error is raised, manually execute `docker start clair_clair`.


### Docker

This method assumes you already have a [PostgreSQL 9.4+] database running.
This is the recommended method for production deployments.

[PostgreSQL 9.4+]: http://postgresql.org

```sh
$ mkdir $HOME/clair_config
$ curl -L https://raw.githubusercontent.com/coreos/clair/v1.2.5/config.example.yaml -o $HOME/clair_config/config.yaml
$ $EDITOR $HOME/clair_config/config.yaml # Add the URI for your postgres database
$ docker run -d -p 6060-6061:6060-6061 -v $HOME/clair_config:/config quay.io/coreos/clair:v1.2.5 -config=/config/config.yaml
```

### Source

To build Clair, you need to latest stable version of [Go] and a working [Go environment].
In addition, Clair requires that [bzr], [rpm], and [xz] be available on the system [$PATH].

[Go]: https://github.com/golang/go/releases
[Go environment]: https://golang.org/doc/code.html
[bzr]: http://bazaar.canonical.com/en
[rpm]: http://www.rpm.org
[xz]: http://tukaani.org/xz
[$PATH]: https://en.wikipedia.org/wiki/PATH_(variable)

```sh
$ go get github.com/coreos/clair
$ go install github.com/coreos/clair/cmd/clair
$ $EDITOR config.yaml # Add the URI for your postgres database
$ ./$GOBIN/clair -config=config.yaml
```

## Documentation

The latest stable documentation can be found [on the CoreOS website]. Documentation for the current branch can be found [inside the Documentation directory][docs-dir] at the root of the project's source code.

[on the CoreOS website]: https://coreos.com/clair/docs/latest/
[docs-dir]: /Documentation

### Architecture at a Glance

![Simple Clair Diagram](img/simple_diagram.png)

### Terminology

- *Image* - a tarball of the contents of a container
- *Layer* - an *appc* or *Docker* image that may or maybe not be dependent on another image
- *Detector* - a Go package that identifies the content, *namespaces* and *features* from a *layer*
- *Namespace* - a context around *features* and *vulnerabilities* (e.g. an operating system)
- *Feature* - anything that when present could be an indication of a *vulnerability* (e.g. the presence of a file or an installed software package)
- *Fetcher* - a Go package that tracks an upstream vulnerability database and imports them into Clair

### Vulnerability Analysis

There are two major ways to perform analysis of programs: [Static Analysis] and [Dynamic Analysis].
Clair has been designed to perform *static analysis*; containers never need to be executed.
Rather, the filesystem of the container image is inspected and *features* are indexed into a database.
By indexing the features of an image into the database, images only need to be rescanned when new *detectors* are added.

[Static Analysis]: https://en.wikipedia.org/wiki/Static_program_analysis
[Dynamic Analysis]: https://en.wikipedia.org/wiki/Dynamic_program_analysis

### Default Data Sources

| Data Source                   | Versions                                               | Format |
|-------------------------------|--------------------------------------------------------|--------|
| [Debian Security Bug Tracker] | 6, 7, 8, unstable                                      | [dpkg] |
| [Ubuntu CVE Tracker]          | 12.04, 12.10, 13.04, 14.04, 14.10, 15.04, 15.10, 16.04 | [dpkg] |
| [Red Hat Security Data]       | 5, 6, 7                                                | [rpm]  |

[Debian Security Bug Tracker]: https://security-tracker.debian.org/tracker
[Ubuntu CVE Tracker]: https://launchpad.net/ubuntu-cve-tracker
[Red Hat Security Data]: https://www.redhat.com/security/data/metrics
[dpkg]: https://en.wikipedia.org/wiki/dpkg
[rpm]: http://www.rpm.org


### Customization

The major components of Clair are all programmatically extensible in the same way Go's standard [database/sql] package is extensible.

Custom behavior can be accomplished by creating a package that contains a type that implements an interface declared in Clair and registering that interface in [init()]. To expose the new behavior, unqualified imports to the package must be added in your [main.go], which should then start Clair using `Boot(*config.Config)`.

The following interfaces can have custom implementations registered via [init()] at compile time:

- `Datastore` - the backing storage
- `Notifier` - the means by which endpoints are notified of vulnerability changes
- `Fetcher` - the sources of vulnerability data that is automatically imported
- `MetadataFetcher` - the sources of vulnerability metadata that is automatically added to known vulnerabilities
- `DataDetector` - the means by which contents of an image are detected
- `FeatureDetector` - the means by which features are identified from a layer
- `NamespaceDetector` - the means by which a namespace is identified from a layer

[init()]: https://golang.org/doc/effective_go.html#init
[database/sql]: https://godoc.org/database/sql
[main.go]: https://github.com/coreos/clair/blob/master/cmd/clair/main.go

## Related Links

- [Talk](https://www.youtube.com/watch?v=PA3oBAgjnkU) and [Slides](https://docs.google.com/presentation/d/1toUKgqLyy1b-pZlDgxONLduiLmt2yaLR0GliBB7b3L0/pub?start=false&loop=false&slide=id.p) @ ContainerDays NYC 2015
- [Quay](https://quay.io): the first container registry to integrate with Clair
- [Dockyard](https://github.com/containerops/dockyard): an open source container registry with Clair integration
