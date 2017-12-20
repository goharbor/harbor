<!--[metadata]>
+++
title = "Notary Command Reference"
description = "Notary command reference"
keywords = ["docker, notary, command, notary command, reference"]
[menu.main]
parent="mn_notary"
weight=99
+++
<![end-metadata]-->

# Notary Command Reference

## Terminology Reference
1. **GUN**: Notary uses Globally Unique Names (GUNs) to identify trusted collections.  Note that `<GUN>` can take on arbitrary values, but when working with Docker Content Trust they are structured like `<domain>/<account>/<repository>`.  For example `docker.io/library/alpine` for the [Docker Library Alpine image](https://hub.docker.com/r/library/alpine).
2. **Target**: A target refers to a file in a to be distributed as part of a trusted collection.  Target files are opaque to the framework. Whether target files are packages containing multiple files, single text files, or executable binaries is irrelevant to Notary.
3. **Trusted Collection**: A trusted collection is a set of target files of interest.
4. **Notary Repository**: A notary repository refers to the set of signed metadata files that describe a trusted collection.
5. **Key roles**:
    - **Root**: The root role delegates trust to specific keys trusted for all other top-level roles used in the system.
    - **Targets**: The targets role's signature indicates which target files are trusted by clients.
    - **Delegations**: the targets role can delegate full or partial trust to other roles.  Delegating trust means that the targets role indicates another role (that is, another set of keys and the threshold required for trust) is trusted to sign target file metadata.
    - **Snapshot**:  The snapshot role signs a metadata file that provides information about the latest version of all of the other metadata on the trusted collection (excluding the timestamp file, discussed below).
    - **Timestamp**:  To prevent an adversary from replaying an out-of-date signed metadata file whose signature has not yet expired, an automated process periodically signs a timestamped statement containing the hash of the snapshot file.

To read further about the framework Notary implements, check out [The Update Framework](https://github.com/theupdateframework/tuf/blob/develop/docs/tuf-spec.txt)

## Command Reference

### Set up Notary CLI

Once you install the Notary CLI client, you can use it to manage your signing
keys, authorize other team members to sign content, and rotate the keys if
a private key has been compromised.

When using the Notary CLI client you need to specify where the URL of the Notary server
you want to communicate is with the `-s` flag, and where to store the private keys and cache for
the CLI client with the `-d` flag.  There are also fields in the [client configuration](/reference/client-config.md) to set these variables.

```bash
# Create an alias to always have the notary client talking to the right notary server
$ alias notary="notary -s <notary_server_url> -d <notary_cache_directory>
```

When working Docker Content Trust, it is important to specify notary's client cache as `~/.docker/trust`.  Also, Docker Hub provides its own Notary server located at `https://notary.docker.io`, which contains trust data for many images including official images, though you are welcome to use your own notary server.

## Initializing a trusted collection

Notary can initialize a trusted collection with the `notary init` command:
```bash
$ notary init <GUN>
```

This command will generate targets and snapshot keys locally for the trusted collection, and try to locate a root key to use in the specified notary client cache (from `-d` or the config).
If notary cannot find a root key, it will generate one.  For all keys, notary will also prompt for a passphrase to encrypt the private key material at rest.

If you'd like to initialize your trusted collection with a specific root key, there is a flag to provide it.  Notary will require that the key is encrypted:
```bash
$ notary init <GUN> --rootkey <key_file>
```

Note that you will have to run a publish after this command for it to take effect, because the Notary CLI client will create staged changes to initialize the trusted collection that have not yet been pushed to a notary server.
```bash
$ notary publish <GUN>
```

## Manage staged changes

The Notary CLI client stages changes before publishing them to the server.
These changes are staged locally in files such that the client can request the latest updates from the notary server, attempt to apply the staged changes on top of the new updates (like a `git rebase`), and then finally publish if the changes are still valid.

You can view staged changes with `notary status` and unstage them with `notary reset`:

```bash
# Check what changes are staged
$ notary status <GUN>

# Unstage a specific change
$ notary reset <GUN> -n 0

# Alternatively, reset all changes
$ notary reset <GUN> --all
```

When you're ready to publish your changes to the Notary server, run:

```bash
$ notary publish <GUN>
```

## Auto-publish changes

Instead of manually running `notary publish` after each command, you can use the `-p` flag to auto-publish the changes from that command.
For example:

```bash
$ notary init -p <GUN>
```

The remainder of this reference will include the `-p` auto-publish flag where it can be used, though it is optional and can be replaced with following each command with a `notary publish`.

## Adding and removing trust data

Users can sign content into a notary trusted collection by running:
```bash
$ notary add -p <GUN> <target_name> <target_file>
```

In the above command, the `<target_name>` corresponds to the name we want to associate the `<target_file>` with in the trusted collection. Notary will sign the hash of the `<target_file>` into its trusted collection.
Instead of adding a target by file, you can specify a hash and byte size directly:
```bash
$ notary addhash -p <GUN> <target_name> <byte_size> --sha256 <sha256Hash>
```

To check that your trust data was published successfully to the notary server, you can run:
```bash
$ notary list <GUN>
```

To remove targets from a trusted collection, you can run:
```bash
$ notary remove -p <GUN> <target_name>
```

## Delete trust data

Users can remove all notary signed data for a trusted collection by running:

```bash
$ notary delete <GUN> --remote
```

If you don't include the `--remote` flag, Notary deletes local cached content
but will not delete data from the Notary server.

## Change the passphrase for a key

The Notary CLI client manages the keys used to sign the trusted collection. These keys are encrypted at rest.
To list all the keys managed by the Notary CLI client, run:

```bash
$ notary key list
```

To change the passphrase used to encrypt one of the keys, run:

```bash
$ notary key passwd <key_id>
```

## Rotate keys

If one of the private keys is compromised you can rotate that key, so that
content that was signed with those keys stop being trusted.

For keys that are kept offline and managed by the Notary CLI client, such the
keys with the root, targets, and snapshot roles, you can rotate them with:

```bash
$ notary key rotate <GUN> <key_role>
```

The Notary CLI client generates a new key for the role you specified, and
prompts you for a passphrase to encrypt it.
Then you're prompted for the passphrase for the key you're rotating, and if it
is correct, the Notary CLI client immediately contacts the Notary server to auto-publish the
change, no `-p` flag needed.

After a rotation, all previously existing keys for the specified role are replaced with the new key.

You can also rotate keys that are stored in the Notary server, such as the keys
with the snapshot or timestamp role. To do this, use the `-r` flag:

```bash
$ notary key rotate <GUN> <key_role> -r
```

## Importing and exporting keys

Notary can import keys that are already in a PEM format:
```bash
$ notary key import <pemfile> --role <key_role> --gun <key_gun>
```

The `--role` and `--gun` flags can be provided to specify a role and GUN to import the key to if that information is not already included in PEM headers.
Note that for root and delegation keys, the `--gun` flag is ignored because these keys can be shared across GUNs.
If no `--role` or `--gun` is given, notary will assume that the key is to be used for a delegation role, which will appear as a `delegation` key in commands such as `notary key list`.


Moreover, it's possible for notary to import multiple keys contained in one PEM file, each separated into separate PEM blocks.

For each key it attempts to import, notary will prompt for a passphrase so that the key can be encrypted at rest.

Notary can also export all of its encrypted keys, or individually by key ID or GUN:
```bash
# export all my notary keys to a file
$ notary key export -o exported_keys.pem

# export a single key by ID
$ notary key export --key <keyID> -o exported_keys.pem

# export multiple keys by ID
$ notary key export --key <keyID1> --key <keyID2> -o exported_keys.pem

# export all keys for a GUN
$ notary key export --gun <GUN> -o exported_keys.pem

# export all keys for multiple GUNs
$ notary key export --gun <GUN1> --gun <GUN2> -o exported_keys.pem
```
When exporting multiple keys, all keys are outputted to a single PEM file in individual blocks. If the output flag `-o` is omitted, the PEM blocks are outputted to STDOUT.

## Manage keys for delegation roles

To delegate content signing to other users without sharing the targets key, retrieve a x509 certificate for that user and run:

```bash
$ notary delegation add -p <GUN> targets/<role> user.pem user2.pem --all-paths
```

The delegated user can then import the private key for that certificate keypair (using `notary key import`) and use it for signing.

The `--all-paths` flag allows the delegation role to sign content into any target name.  To restrict this, you can provide path prefixes with the `--paths` flag instead.  For example:
```bash
$ notary delegation add -p <GUN> targets/<role> user.pem user2.pem --paths tmp/ --paths users/
```
In the above example, the delegation would be allowed to sign targets prefixed by `tmp/` and `users/` (ex: `tmp/file`, `users/file`, but not `file`)

It's possible to add multiple certificates at once for a role:
```bash
$ notary delegation add -p <GUN> targets/<role> --all-paths user1.pem user2.pem user3.pem
```

You can also remove keys from a delegation role, such that those keys can no longer sign targets into the delegation role:

```bash
# Remove the given keys from a delegation role
$ notary delegation remove -p <GUN> targets/<role> <keyID1> <keyID2>

# Alternatively, you can remove keys from all delegation roles, in case of delegation key compromise
$ notary delegation purge <GUN> --key <keyID1> --key <keyID2>
```

## Managing targets in delegation roles

We can specify which delegation roles to sign content into by using the `--roles` flag.  This also applies to `notary addhash` and `notary remove`.
Without the `--roles` flag, notary will attempt to operate on the base `targets` role:
```bash
# Add content from a target file to a specific delegation role
$ notary add -p <GUN> <target_name> <target_file> --roles targets/<role>

# Add content by hash to a specific delegation role
$ notary addhash -p <GUN> <target_name> <byte_size> --sha256 <sha256Hash> --roles targets/<role>

# Remove content from a specific delegation role
$ notary remove -p <GUN> <target_name> <target_file> --roles targets/<role>
```

Similarly, we can list all targets and prefer to show certain delegation roles' targets first with the `--roles` flag.
If we do not specify a `--role` flag in `notary list`, we will prefer to show targets signed into the base `targets` role, and these will shadow other targets signed into delegation roles with the same target name:
```bash
# Prefer to show targets from one specific role
$ notary list <GUN> --roles targets/<role>

# Prefer to show targets from this list of roles
$ notary list <GUN> --roles targets/<role1> --roles targets/<role2>
```

## Witnessing delegations

Notary can mark a delegation role for re-signing without adding any additional content:

```bash
$ notary witness -p <GUN> targets/<role>
```

This is desirable if you would like to sign a delegation role's existing contents with a new key.

It's possible that this could be useful for auditing, but moreover it can help recover a delegation role that may have become invalid.
For example: Alice last updated delegation `targets/qa`, but Alice since left the company and an administrator has removed her delegation key from the repo.
Now delegation `targets/qa` has no valid signatures, but another signer in that delegation role can run `notary witness targets/qa` to sign off on the existing contents, provided it is still trusted content.

## Troubleshooting

Notary CLI has a `-D` flag that you can use to increase the logging level. You
can use this for troubleshooting.

Usually most problems are fixed by ensuring you're communicating with the
correct Notary server, using the `-s` flag, and that you're using the correct
directory where your private keys are stored, with the `-d` flag.

If you are receiving this error:
```bash
* fatal: Get <URL>/v2/: x509: certificate signed by unknown authority
```
The Notary CLI must be configured to trust the root certificate authority of the server it is communicating with.
This can be configured by specifying the root certificate with the `--tlscacert` flag or by editing the Notary client configuration file.  Additionally, you can add the root certificate to your system CAs.