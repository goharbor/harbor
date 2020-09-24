<img src="docs/images/notary-blk.svg" alt="Notary" width="400px"/>

[![GoDoc](https://godoc.org/github.com/theupdateframework/notary?status.svg)](https://godoc.org/github.com/theupdateframework/notary)
[![Circle CI](https://circleci.com/gh/theupdateframework/notary/tree/master.svg?style=shield)](https://circleci.com/gh/theupdateframework/notary/tree/master) [![CodeCov](https://codecov.io/github/theupdateframework/notary/coverage.svg?branch=master)](https://codecov.io/github/theupdateframework/notary) [![GoReportCard](https://goreportcard.com/badge/theupdateframework/notary)](https://goreportcard.com/report/github.com/theupdateframework/notary)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Ftheupdateframework%2Fnotary.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Ftheupdateframework%2Fnotary?ref=badge_shield)

# Notice

The Notary project has officially been accepted in to the Cloud Native Computing Foundation (CNCF).
It has moved to https://github.com/theupdateframework/notary. Any downstream consumers should update
their Go imports to use this new location, which will be the canonical location going forward.

We have moved the repo in GitHub, which will allow existing importers to continue using the old
location via GitHub's redirect.

# Overview

The Notary project comprises a [server](cmd/notary-server) and a [client](cmd/notary) for running and interacting
with trusted collections. See the [service architecture](docs/service_architecture.md) documentation
for more information.

Notary aims to make the internet more secure by making it easy for people to
publish and verify content. We often rely on TLS to secure our communications
with a web server which is inherently flawed, as any compromise of the server
enables malicious content to be substituted for the legitimate content.

With Notary, publishers can sign their content offline using keys kept highly
secure. Once the publisher is ready to make the content available, they can
push their signed trusted collection to a Notary Server.

Consumers, having acquired the publisher's public key through a secure channel,
can then communicate with any Notary server or (insecure) mirror, relying
only on the publisher's key to determine the validity and integrity of the
received content.

## Goals

Notary is based on [The Update Framework](https://www.theupdateframework.com/), a secure general design for the problem of software distribution and updates. By using TUF, Notary achieves a number of key advantages:

* **Survivable Key Compromise**: Content publishers must manage keys in order to sign their content. Signing keys may be compromised or lost so systems must be designed in order to be flexible and recoverable in the case of key compromise. TUF's notion of key roles is utilized to separate responsibilities across a hierarchy of keys such that loss of any particular key (except the root role) by itself is not fatal to the security of the system.
* **Freshness Guarantees**: Replay attacks are a common problem in designing secure systems, where previously valid payloads are replayed to trick another system. The same problem exists in the software update systems, where old signed can be presented as the most recent. Notary makes use of timestamping on publishing so that consumers can know that they are receiving the most up to date content. This is particularly important when dealing with software update where old vulnerable versions could be used to attack users.
* **Configurable Trust Thresholds**: Oftentimes there are a large number of publishers that are allowed to publish a particular piece of content. For example, open source projects where there are a number of core maintainers. Trust thresholds can be used so that content consumers require a configurable number of signatures on a piece of content in order to trust it. Using thresholds increases security so that loss of individual signing keys doesn't allow publishing of malicious content.
* **Signing Delegation**: To allow for flexible publishing of trusted collections, a content publisher can delegate part of their collection to another signer. This delegation is represented as signed metadata so that a consumer of the content can verify both the content and the delegation.
* **Use of Existing Distribution**: Notary's trust guarantees are not tied at all to particular distribution channels from which content is delivered. Therefore, trust can be added to any existing content delivery mechanism.
* **Untrusted Mirrors and Transport**: All of the notary metadata can be mirrored and distributed via arbitrary channels.

## Security

See Notary's [service architecture docs](docs/service_architecture.md#threat-model) for more information about our threat model, which details the varying survivability and severities for key compromise as well as mitigations.

Notary's last security audit was on July 31, 2015 by NCC ([results](docs/resources/ncc_docker_notary_audit_2015_07_31.pdf)).

Any security vulnerabilities can be reported to security@docker.com.

# Getting started with the Notary CLI

Get the Notary Client CLI binary from [the official releases page](https://github.com/theupdateframework/notary/releases) or you can [build one yourself](#building-notary).
The version of the Notary server and signer should be greater than or equal to Notary CLI's version to ensure feature compatibility (ex: CLI version 0.2, server/signer version >= 0.2), and all official releases are associated with GitHub tags.

To use the Notary CLI with Docker hub images, have a look at Notary's
[getting started docs](docs/getting_started.md).

For more advanced usage, see the
[advanced usage docs](docs/advanced_usage.md).

To use the CLI against a local Notary server rather than against Docker Hub:

1. Ensure that you have [docker and docker-compose](http://docs.docker.com/compose/install/) installed.
1. `git clone https://github.com/theupdateframework/notary.git` and from the cloned repository path,
    start up a local Notary server and signer and copy the config file and testing certs to your
    local Notary config directory:

    ```sh
    $ docker-compose build
    $ docker-compose up -d
    $ mkdir -p ~/.notary && cp cmd/notary/config.json cmd/notary/root-ca.crt ~/.notary
    ```

1. Add `127.0.0.1 notary-server` to your `/etc/hosts`, or if using docker-machine,
    add `$(docker-machine ip) notary-server`).

You can run through the examples in the
[getting started docs](docs/getting_started.md) and
[advanced usage docs](docs/advanced_usage.md), but
without the `-s` (server URL) argument to the `notary` command since the server
URL is specified already in the configuration, file you copied.

You can also leave off the `-d ~/.docker/trust` argument if you do not care
to use `notary` with Docker images.


## Building Notary

Note that Notary's [latest stable release](https://github.com/theupdateframework/notary/releases) is at the head of the
[releases branch](https://github.com/theupdateframework/notary/tree/releases).  The master branch is the development
branch and contains features for the next release.

Prerequisites:

- Go >= 1.7.1
    - Fedora: `dnf install golang`
- libtool development headers installed
    - Ubuntu: `apt-get install libltdl-dev`
    - CentOS/RedHat: `yum install libtool-ltdl-devel`
    - Fedora: `dnf install libtool-ltdl-devel`
    - Mac OS ([Homebrew](http://brew.sh/)): `brew install libtool`

Set [```GOPATH```](https://golang.org/doc/code.html#GOPATH). Then, run:

```bash
$ go get github.com/theupdateframework/notary
# build with pcks11 support by default to support yubikey
$ go install -tags pkcs11 github.com/theupdateframework/notary/cmd/notary
$ notary
```

To build the server and signer, run `docker-compose build`.


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Ftheupdateframework%2Fnotary.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Ftheupdateframework%2Fnotary?ref=badge_large)
