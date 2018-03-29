<!--[metadata]>
+++
title = "Understand the service architecture"
description = "How the three requisite notary components interact"
keywords = ["docker, notary, notary-client, docker content trust, content trust, notary-server, notary server, notary-signer, notary signer, notary architecture"]
[menu.main]
parent="mn_notary"
weight=3
+++
<![end-metadata]-->


# Understand the Notary service architecture

On this page, you get an overview of the Notary service architecture.

## Brief overview of TUF keys and roles

This document assumes familiarity with
<a href="https://www.theupdateframework.com/" target="_blank">The Update Framework</a>,
but here is a brief recap of the TUF roles and corresponding key hierarchy:

<center><img src="https://cdn.rawgit.com/docker/notary/09f81717080f53276e6881ece57cbbbf91b8e2a7/docs/images/key-hierarchy.svg" alt="TUF Key Hierarchy" width="400px"/></center>

- The root key is the root of all trust. It signs the
  <a href="https://github.com/theupdateframework/tuf/blob/1bed3e09a478c2c918ffbff10b9118f6e52ee129/docs/tuf-spec.txt#L489" target="_blank">root metadata file</a>,
  which lists the IDs of the root, targets, snapshot, and timestamp public keys.
  Clients use these public keys to verify the signatures on all the metadata files
  in the repository. This key is held by a collection owner, and should be kept offline
  and safe, more so than any other key.

- The snapshot key signs the
  <a href="https://github.com/theupdateframework/tuf/blob/1bed3e09a478c2c918ffbff10b9118f6e52ee129/docs/tuf-spec.txt#L604" target="_blank">snapshot metadata file</a>,
  which enumerates the filenames, sizes, and hashes of the root,
  targets, and delegation metadata files for the collection. This file is used to
  verify the integrity of the other metadata files. The snapshot key is held by
  either a collection owner/administrator, or held by the Notary service to facilitate
  [signing by multiple collaborators via delegation roles](advanced_usage.md#working-with-delegation-roles).

- The timestamp key signs the
  <a href="https://github.com/theupdateframework/tuf/blob/1bed3e09a478c2c918ffbff10b9118f6e52ee129/docs/tuf-spec.txt#L827" target="_blank">timestamp metadata file</a>,
  which provides freshness guarantees for the collection by having the shortest expiry time of any particular
  piece of metadata and by specifying the filename, size, and hash of the most recent
  snapshot for the collection. It is used to verify the integrity of the snapshot
  file. The timestamp key is held by the Notary service so the timestamp can be
  automatically re-generated when it is requested from the server, rather than
  require that a collection owner come online before each timestamp expiry.

- The targets key signs the
  <a href="https://github.com/theupdateframework/tuf/blob/1bed3e09a478c2c918ffbff10b9118f6e52ee129/docs/tuf-spec.txt#L678" target="_blank">targets metadata file</a>,
  which lists filenames in the collection, and their sizes and respective
  <a href="https://en.wikipedia.org/wiki/Cryptographic_hash_function" target="_blank">hashes</a>.
  This file is used to verify the integrity of some or all of the actual contents of the repository.
  It is also used to
  [delegate trust to other collaborators via delegation roles](advanced_usage.md#working-with-delegation-roles).
  The targets key is held by the collection owner or administrator.

- Delegation keys sign
  <a href="https://github.com/theupdateframework/tuf/blob/1bed3e09a478c2c918ffbff10b9118f6e52ee129/docs/tuf-spec.txt#L678" target="_blank">delegation metadata files</a>,
  which lists filenames in the collection, and their sizes and respective
  <a href="https://en.wikipedia.org/wiki/Cryptographic_hash_function" target="_blank">hashes</a>.
  These files are used to verify the integrity of some or all of the actual contents of the repository.
  They are also used to [delegate trust to other collaborators via lower level delegation roles](
  advanced_usage.md#working-with-delegation-roles).
  Delegation keys are held by anyone from the collection owner or administrator to
  collection collaborators.

## Architecture and components

Notary clients pull metadata from one or more (remote) Notary services. Some
Notary clients will push metadata to one or more Notary services.

A Notary service consists of a Notary server, which stores and updates the
signed
<a href="https://github.com/theupdateframework/tuf/blob/1bed3e09a478c2c918ffbff10b9118f6e52ee129/docs/tuf-spec.txt#L348">TUF metadata files</a>
for multiple trusted collections in an associated database, and a Notary signer, which
stores private keys for and signs metadata for the Notary server. The following
diagram illustrates this architecture:

![Notary Service Architecture Diagram](https://cdn.rawgit.com/docker/notary/09f81717080f53276e6881ece57cbbbf91b8e2a7/docs/images/service-architecture.svg)

Root, targets, and (sometimes) snapshot metadata are generated and signed by
clients, who upload the metadata to the Notary server. The server is
responsible for:

- ensuring that any uploaded metadata is valid, signed, and self-consistent
- generating the timestamp (and sometimes snapshot) metadata
- storing and serving to clients the latest valid metadata for any trusted collection

The Notary signer is responsible for:

- storing the private signing keys
  <a href="https://tools.ietf.org/html/draft-ietf-jose-json-web-algorithms-31#section-4.4" target="_blank">wrapped</a>
  and <a href="https://tools.ietf.org/html/draft-ietf-jose-json-web-algorithms-31#section-4.8" target="_blank">encrypted</a>
  using <a href="https://github.com/dvsekhvalnov/jose2go" target="_blank">Javascript Object Signing and Encryption</a> in a database separate from the
  Notary server database
- performing signing operations with these keys whenever the Notary server requests

## Example client-server-signer interaction

The following diagram illustrates the interactions between the Notary client,
server, and signer:

![Notary Service Sequence Diagram](https://cdn.rawgit.com/docker/notary/27469f01fe244bdf70f34219616657b336724bc3/docs/images/metadata-sequence.svg)

1. Notary server optionally supports authentication from clients using
   <a href="http://jwt.io/" target="_blank">JWT</a> tokens. This requires an authorization server that
   manages access controls, and a cert bundle from this authorization server
   containing the public key it uses to sign tokens.

    If token authentication is enabled on Notary server, then any connecting
    client that does not have a token will be redirected to the authorization
    server.

    Please see the docs for [Docker Registry v2 authentication](
    https://github.com/docker/distribution/blob/master/docs/spec/auth/token.md)
    for more information.

2. The client will log in to the authorization server via basic auth over HTTPS,
   obtain a bearer token, and then present the token to Notary server on future
   requests.

3. When clients uploads new metadata files, Notary server checks them against
   any previous versions for conflicts, and verifies the signatures, checksums,
   and validity of the uploaded metadata.

4. Once all the uploaded metadata has been validated, Notary server
   generates the timestamp (and maybe snapshot) metadata. It sends this
   generated metadata to the Notary signer to be signed.

5. Notary signer retrieves the necessary encrypted private keys from its database
   if available, decrypts the keys, and uses them to sign the metadata. If
   successful, it sends the signatures back to Notary server.

6. Notary server is the source of truth for the state of a trusted collection of
   data, storing both client-uploaded and server-generated metadata in the TUF
   database. The generated timestamp and snapshot metadata certify that the
   metadata files the client uploaded are the most recent for that trusted collection.

    Finally, Notary server will notify the client that their upload was successful.

7. The client can now immediately download the latest metadata from the server,
   using the still-valid bearer token to connect. Notary server only needs to
   obtain the metadata from the database, since none of the metadata has expired.

    In the case that the timestamp has expired, Notary server would go through
    the entire sequence where it generates a new timestamp, request Notary signer
    for a signature, stores the newly signed timestamp in the database. It then
    sends this new timestamp, along with the rest of the stored metadata, to the
    requesting client.


## Threat model

Both the server and the signer are potential attack vectors against all users
of the Notary service. Client keys are also a potential attack vector, but
not necessarily against all collections at a time. This section
discusses how our architecture is designed to deal with compromises.

### Notary server compromise

In the event of a Notary server compromise, an attacker would have direct access to
the metadata stored in the database as well as well as access to the credentials
used to communicate with Notary signer, and therefore, access to arbitrary signing
operations with any key the Signer holds.

- **Denial of Service** - An attacker could reject client requests and corrupt
    or delete metadata from the database, thus preventing clients from being
    able to download or upload metadata.

- **Malicious Content** - An attacker can create, store, and serve arbitrary
    metadata content for one or more trusted collections. However, they do not have
    access to any client-side keys, such as root, targets, and potentially the
    snapshot keys for the existing trusted collections.

    Only clients who have never seen the trusted collections, and who do not have any
    form of pinned trust, can be tricked into downloading and
    trusting the malicious content for these trusted collections.

    Clients that have previously interacted with any trusted collection, or that have
    their trust pinned to a specific certificate for the collections will immediately
    detect that the content is malicious and would not trust any root, targets,
    or (maybe) snapshot metadata for these collections.

- **Rollback, Freeze, Mix and Match** - The attacker can request that
    the Notary signer sign any arbitrary timestamp (and maybe snapshot) metadata
    they want. Attackers can launch a freeze attack, and, depending on whether
    the snapshot key is available, a mix-and-match attack up to the expiration
    of the targets file.

    Clients both with and without pinned trust would be vulnerable to these
    attacks, so long as the attacker ensures that the version number of their
    malicious metadata is higher than the version number of the most recent
    good metadata that any client may have.

 Note that the timestamp and snapshot keys cannot be compromised in a server-only
 compromise, so a key rotation would not be necessary. Once the Server
 compromise is mitigated, an attacker will not be
 able to generate valid timestamp or snapshot metadata and serve them on a
 malicious mirror, for example.

### Notary signer compromise

In the event of a Notary signer compromise, an attacker would have access to
all the (timestamp and snapshot) private keys stored in a database.
If the keys are stored in an HSM, they would have the ability to sign arbitrary
content with, and to delete, the keys in the HSM, but not to exfiltrate the
private material.

- **Denial of Service** - An attacker could reject all Notary server requests
  and corrupt or delete keys from the database (or even delete keys from an
  HSM), and thus prevent Notary servers from being able to sign generated
  timestamps or snapshots.

- **Key Compromise** - If the Notary signer uses a database as its backend,
  an attacker can exfiltrate all the (timestamp and snapshot) private material.
  Note that the capabilities of an attacker are the same as of a Notary server
  compromise in terms of signing arbitrary metadata, with the important detail
  that in this particular case key rotations will be necessary to recover from
  the attack.

### Notary client keys and credentials compromise

The security of keys held and administered by users depends on measures taken by
the users. If the Notary Client CLI was used to create them, then they are password
protected and the Notary CLI does not provide options to export them in
plaintext.

It is up to the user to choose an appropriate password, and to protect their key
from offline brute-force attacks.

The severity of the compromise of a trust collection owner/administrator's
decrypted key depends on the type and combination of keys that were compromised
(e.g. the snapshot key and targets key, or just the targets key).

#### Possible attacks given the credentials compromised:

- **Decrypted Delegation Key, only**

    | Keys compromised | Malicious Content | Rollback, Freeze, Mix and Match | Denial of Service |
    |------------------|-------------------|---------------------------------|-------------------|
    | Delegation key   | no                | no                              | no                |


- **Decrypted Delegation Key + Notary Service write-capable credentials**

    | Keys compromised | Malicious Content | Rollback, Freeze, Mix and Match | Denial of Service |
    |------------------|-------------------|---------------------------------|-------------------|
    | Delegation key   | limited, maybe*   | limited, maybe*                 | limited, maybe*   |

  *If the Notary Service holds the snapshot key and the attacker has Notary Service
  write credentials, then they have effective access to the snapshot and timestamp
  keys because the server will generate and sign the snapshot and timestamp for them.

  An attacker can add malicious content, remove legitimate content from a collection, and
  mix up the targets in a collection, but only within the particular delegation
  roles that the key can sign for. Depending on the restrictions on that role,
  they may be restricted in what type of content they can modify. They may also
  add or remove the capabilities of other delegation keys below it on the key hierarchy
  (e.g. if `DelegationKey2` in the above key hierarchy were compromised, it would only be
  able to modify the capabilities of `DelegationKey4` and `DelegationKey5`).

- **Decrypted Delegation Key + Decrypted Snapshot Key, only**

    | Keys compromised | Malicious Content | Rollback, Freeze, Mix and Match | Denial of Service |
    |------------------|-------------------|---------------------------------|-------------------|
    | Delegation key <br/> Snapshot key  | no    | no                        | no                |

  The attacker does not have access to the timestamp key, which is always held by the Notary
  Service, and will be unable to set up a malicious mirror.

- **Decrypted Delegation Key + Decrypted Snapshot Key + Notary Service write-capable credentials**

    | Keys compromised | Malicious Content | Rollback, Freeze, Mix and Match | Denial of Service |
    |------------------|-------------------|---------------------------------|-------------------|
    | Delegation key <br/> Snapshot key  | limited   | limited               | limited           |

  The Notary Service always holds the timestamp key. If the attacker has Notary Service
  write credentials, then they have effective access to the timestamp key because the server
  will generate and sign the timestamp for them.

  An attacker can add malicious content, remove legitimate content from a collection, and
  mix up the targets in a collection, but only within the particular delegation
  roles that the key can sign for. Depending on the restrictions on that role,
  they may be restricted in what type of content they can modify. They may also
  add or remove the capabilities of other delegation keys below it on the key hierarchy
  (e.g. if `DelegationKey2` in the above key hierarchy were compromised, it would only be
  able to modify the capabilities of `DelegationKey4` and `DelegationKey5`).

- **Decrypted Targets Key, only**

    | Keys compromised | Malicious Content | Rollback, Freeze, Mix and Match | Denial of Service |
    |------------------|-------------------|---------------------------------|-------------------|
    | Targets key      | no                | no                              | no                |

- **Decrypted Targets Key + Notary Service write-capable credentials**

    | Keys compromised | Malicious Content | Rollback, Freeze, Mix and Match | Denial of Service |
    |------------------|-------------------|---------------------------------|-------------------|
    | Targets key      | maybe*            | maybe*                          | limited, maybe*   |

  *If the Notary Service holds the snapshot key and the attacker has Notary Service
  write credentials, then they have effective access to the snapshot and timestamp
  keys because the server will generate and sign the snapshot and timestamp for them.

  An attacker can add any malicious content, remove any legitimate content from a
  collection, and mix up the targets in a collection. They may also add or remove
  the capabilities of any top level delegation key or role (e.g. `Delegation1`,
  `Delegation2`, and `Delegation3` in the key hierarchy diagram). If they remove
  the roles entirely, they'd break the trust chain to the lower delegation roles
  (e.g. `Delegation4`, `Delegation5`).

- **Decrypted Targets Key + Decrypted Snapshot Key, only**

    | Keys compromised | Malicious Content | Rollback, Freeze, Mix and Match | Denial of Service |
    |------------------|-------------------|---------------------------------|-------------------|
    | Targets key <br/> Snapshot key     | no    | no                        | no                |

  The attacker does not have access to the timestamp key, which is always held by the Notary
  Service, and will be unable to set up a malicious mirror.

- **Decrypted Targets Key + Decrypted Snapshot Key + Notary Service write-capable credentials**

    | Keys compromised | Malicious Content | Rollback, Freeze, Mix and Match | Denial of Service |
    |------------------|-------------------|---------------------------------|-------------------|
    | Targets key <br/> Snapshot key       | yes   | yes                     | limited           |

  The Notary Service always holds the timestamp key. If the attacker has Notary Service
  write credentials, then they have effective access to the timestamp key because the server
  will generate and sign the timestamp for them.

  An attacker can add any malicious content, remove any legitimate content from a
  collection, and mix up the targets in a collection. They may also add or remove
  the capabilities of any top level delegation key or role (e.g. `Delegation1`,
  `Delegation2`, and `Delegation3` in the key hierarchy diagram). If they remove
  the roles entirely, they'd break the trust chain to the lower delegation roles
  (e.g. `Delegation4`, `Delegation5`).

- **Decrypted Root Key + none or any combination of decrypted keys, only**

    | Keys compromised | Malicious Content | Rollback, Freeze, Mix and Match | Denial of Service |
    |------------------|-------------------|---------------------------------|-------------------|
    | All keys         | yes               | yes                             | yes               |

  No other keys are needed, since the attacker can just any rotate or all of them to ones that they
  generate. With these keys, they can set up a mirror to serve malicious data - any malicious data
  at all, given that they have access to all the keys.

- **Decrypted Root Key + none or any combination of decrypted keys + Notary Service write-capable credentials**

    | Keys compromised | Malicious Content | Rollback, Freeze, Mix and Match | Denial of Service |
    |------------------|-------------------|---------------------------------|-------------------|
    | All keys         | yes               | yes                             | yes               |

  *If the Notary Service holds the snapshot key and the attacker has Notary Service
  write credentials, then they won't even have to rotate the snapshot and timestamp
  keys because the server will generate and sign the snapshot and timestamp for them.

#### Mitigations

If a root key compromise is detected, the root key holder should contact
whomever runs the notary service to manually reverse any malicious changes to
the repository, and immediately rotate the root key. This will create a fork
of the repository history, and thus break existing clients who have downloaded
any of the malicious changes.

If a targets key compromise is detected, the root key holder
must rotate the compromised key and push a clean set of targets using the new key.

If a delegations key compromise is detected, a higher level key (e.g. if
`Delegation4` were compromised, then `Delegation2`; if
`Delegation2` were compromised, then the `Targets` key)
holder must rotate the compromised key, and push a clean set of targets using the new key.

If a Notary Service credential compromise is detected, the credentials should be
changed immediately.

## Related information

* [Run a Notary service](running_a_service.md)
* [Notary configuration files](reference/index.md)
