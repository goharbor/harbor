<!--[metadata]>
+++
title = "Notary Changelog"
description = "Notary release changelog"
keywords = ["docker, notary, changelog, notary changelog, notary releases, releases, notary versions, versions"]
[menu.main]
parent="mn_notary"
weight=99
+++
<![end-metadata]-->

# Changelog

## v0.3
#### 5/11/2016
Implements root key and certificate rotation, as well as trust pinning configurations to specify known good key IDs and CAs to replace TOFU.
Additional improvements and fixes to notary internals, and RethinkDB support.

> Detailed release notes can be found here:
<a href="https://github.com/docker/notary/releases/tag/v0.3.0" target="_blank">v0.3 release notes</a>.

## v0.2
#### 2/24/2016
Adds support for
<a href="https://github.com/theupdateframework/tuf/blob/1bed3e09a478c2c918ffbff10b9118f6e52ee129/docs/tuf-spec.txt#L387" target="_blank">delegation
roles</a> in TUF.
Delegations allow for easier key management amongst collaborators in a notary trusted collection, and fine-grained permissions on what content each delegate is allowed to modify and sign.
This version also supports managing the snapshot key on notary server, which should be used when enabling delegations on a trusted collection.
Moreover, this version also adds more key management functionality to the notary CLI, and changes the docker-compose development configuration to use the official MariaDB image.

> Detailed release notes can be found here:
<a href="https://github.com/docker/notary/releases/tag/v0.2.0" target="_blank">v0.2 release notes</a>.

## v0.1
#### 11/15/2015
Initial notary non-alpha release.
Implements The Update Framework (TUF) with root, targets, snapshot, and timestamp roles to sign and verify content of a trusted collection.

> Detailed release notes can be found here:
<a href="https://github.com/docker/notary/releases/tag/v0.1" target="_blank">v0.1 release notes</a>.
