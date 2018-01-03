<!--[metadata]>
+++
title = "Best Practices for Using Notary"
description = "A set of recommended practices for key management, delegating trust, bootstrapping repos, and more."
keywords = ["docker, Notary, notary-client, docker content trust, content trust", "best practices", "recommended use"]
[menu.main]
parent="mn_notary"
weight=1
+++
<![end-metadata]-->

# Best Practices for Using Docker Notary

This document describes good and recommended practices for using the Notary client.
It includes recommendations on key management, bootstrapping trust for your repositories,
and delegating trust to other signers. You may wish to refer to the [Getting Started](getting_started.md)
and [Advanced Usage](advanced_usage.md) documents for more information on some of the
commands.

## Key Management

There are five primary key roles in Notary. In order of how critical they are to protect, most
to least, they are:

1. Root
2. Targets
3. Delegations
4. Snapshot
5. Timestamp

We will address each of these classifications of key one by one in reverse order. The
Targets and Delegations key will be covered together as they have almost identical
security profiles and use cases.

### Timestamp Key

It is required that the server manages your Timestamp key. This is a convenience measure
as it would be impractical to have a human involved in regularly signing the Timestamp due
to the frequency with which it expires.

### Snapshot Key

You may choose to allow the server to manage your Snapshot key. This behaviour is default
in the Docker Content Trust integration. Typically if you are working on your own, and
publish new content regularly, it is safe for you to manage the Snapshot key yourself.
If however you use Delegations to collaborate with other contributors, you will want to
allow the server to sign snapshots.

We consider holding the Timestamp and Snapshot keys in secured online locations an
acceptable tradeoff between perfect security and usability.

### Targets and Delegations Keys

While compromise of the Timestamp and Snapshot keys allows an attacker to potentially trick
new or very out of date users into trusting older or mixed versions of the The Update Framework (TUF) repository, compromise
of a Delegations key, or the Targets key, allows an attacker to sign in completely arbitrary
content. Therefore it is important to keep these keys secured. However, as they are also
required for adding, updating, and deleting content in your Trusted Collection, they are
more likely to be stored on one or more of your personal computers. The Notary client requests
a password when saving any keys to persistent storage. This password is used as one component
in the generation of a symmetric AES 256 key that is used to encrypt your signing key in CBC
mode. When the signing key is required, you are asked to supply the password so that the
symmetric key can be regenerated to decrypt the signing key.

While you may be tempted to mirror SSH common practices and generate different Delegations
keys for every computer you publish from, the provided key rotation functionalities of TUF
make this unnecessary. However, there is no security weakness in doing so and if you use
a large number of different computers, it may be desirable simply from a key distribution
perspective in that you have fewer systems to update should one key be compromised. 

The Targets key is placed absolutely above Delegations keys because it is ultimately
directly or indirectly, responsible for all Delegations keys permitted to sign content
into the Trusted Collection. Additionally, rotating the Targets key requires access to
the Root key which has much more stringent security requirements, whereas rotating a
Delegation key requires another Delegation key, or the Targets key, depending on the
structure of the Notary repository.

### Root Key

Compromise of any of the other keys can be easily dealt with via a normal key rotation. The Root
key however anchors trust and in the case of a Targets, Snapshot, or Timestamp key compromise,
is used to perform said key rotation. While a Root key compromise can be handled via the special
Root key rotation mechanism, the probability of a Root key compromise should be reduced as much as is possible.
Therefore, we strongly recommend that a Yubikey is used to store a copy of the Root key, and
is itself stored in a secure location like a safe. Furthermore, a paper copy of the Root key
should be stored in a bank vault. Note that when Notary generates a new Root key, regardless
of whether a Yubikey is present, it will always store a password encrypted copy on your local
filesystem specifically so that you can make these types of backups. The `notary key import`
command will allow you to create multiple Yubikeys with the Root key loaded.

Additionally, you may choose whether or not to re-use a Root key for multiple repositories.
Notary will automatically use the first Root key it finds available, or generate a new Root
key if none is found. By ensuring only the Root key you want to create the repository with
is available, i.e. by attaching the appropriate Yubikey, Notary will use this key as the
root key for a new repository.

## Key Rotations

If you follow the appropriate security measures, you should not need to rotate your Root
key until the signing mechanism in use is obsoleted. We currently default to ECDSA 
with the P256 curve and do not expect it to be obsoleted for many years. Currently 
the Root key is published as a self signed x509 certificate with an expiry set 10 years
in the future. Therefore, you will need to rotate your root key within this time frame.
While the root key shouldn't need to be rotated often, it is still beneficial to periodically
require an update to improved or recommended cipher suites. We chose a 10 year expiry as the
default in order to enforce this requirement.

It may however
be desirable to rotate other keys at more frequent intervals due to their increased exposure
to potential compromise. We have no specific recommendations on the period length.

## Expiration Prevention

Keys held online by the Notary signer will be used to re-sign their associated roles
if the current metadata expires. The current default metadata expiry times, which will
be reduced as people become familiar with the management of Notary repositories, are:

* Root - 10 years
* Targets and Delegations - 3 years
* Snapshot - 3 years
* Timestamp - 14 days

If you are publishing content that should not be expired after 3 years, you will need to
ensure you have a mechanism in place to re-sign your content before it expires. If the
Targets role expires, no content in the Trusted Collection will be considered valid by
client. If however a delegation expires, only the items in the subtree rooted at that
delegation will be invalidated.

If your Root file expires, the repository as a whole is no longer trusted and all clients
will fail when interacting with your repository. The Notary client, and the Docker Content
Trust integration, will attempt to re-sign the Root file if it is within 6 months of expiring
when a publish is performed.

Whenever the Notary repository is retrieved and a role other than the timestamp is found 
to be within 6 months of expiring, warnings to be printed to the logs (or the terminal in 
the case of the Notary or Docker CLIs). This is to notify both consumers, and administrators 
that content they were trusting may become invalid due to inactivity.

## Bootstrapping Trust

By default Notary uses a mechanism we have termed TOFUS (Trust On First Use over HTTPS).
This is not ideal but is no worse than typical trust bootstrapping, which would involve
retrieving a publisher's public key from their website, optimally over TLS. However, in
many cases, we wish to use Notary to establish trust over packages we ourselves have
published, or packages from people and organizations we have direct relationships with.
In these instances we can do better than TOFUS as we already have the public key, or have more
secure ways to retrieve it.

We recommend that wherever possible, trust is bootstrapped by pinning specific Root
keys for specific repositories, as described in the [client configuration](reference/client-config.md)
document.