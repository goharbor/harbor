# Changelog

## [v0.6.1](https://github.com/docker/notary/releases/tag/v0.6.0) 04/10/2018
+ Fixed bug where CLI requested admin privileges for all metadata operations, including listing targets on a repo [#1315](https://github.com/theupdateframework/notary/pull/1315)
+ Prevented notary signer from being dumpable or ptraceable in Linux, except in debug mode [#1327](https://github.com/theupdateframework/notary/pull/1327)
+ Bumped JWT dependency to fix potential Invalid Curve Attack on NIST curves within ECDH key management [#1334](https://github.com/theupdateframework/notary/pull/1334)
+ If the home directory cannot be found, log a warning instead of erroring out [#1318](https://github.com/theupdateframework/notary/pull/1318)
+ Bumped go version and various dependencies [#1323](https://github.com/theupdateframework/notary/pull/1323) [#1332](https://github.com/theupdateframework/notary/pull/1332) [#1335](https://github.com/theupdateframework/notary/pull/1335) [#1336](https://github.com/theupdateframework/notary/pull/1336)
+ Various internal and documentation fixes [#1312](https://github.com/theupdateframework/notary/pull/1312) [#1313](https://github.com/theupdateframework/notary/pull/1313) [#1319](https://github.com/theupdateframework/notary/pull/1319) [#1320](https://github.com/theupdateframework/notary/pull/1320) [#1324](https://github.com/theupdateframework/notary/pull/1324) [#1326](https://github.com/theupdateframework/notary/pull/1326) [#1328](https://github.com/theupdateframework/notary/pull/1328) [#1329](https://github.com/theupdateframework/notary/pull/1329) [#1333](https://github.com/theupdateframework/notary/pull/1333)

## [v0.6.0](https://github.com/docker/notary/releases/tag/v0.6.0) 02/28/2018
+ **The project has been moved from https://github.com/docker/notary to https://github.com/theupdateframework/notary, as it has been accepted into the CNCF.  Downstream users should update their go imports.**
+ Removed support for RSA-key exchange ciphers supported by the server and signer and require TLS >= 1.2 for the server and signer. [#1307](https://github.com/theupdateframework/notary/pull/1307)
+ `libykcs11` can be found in several additional locations on Fedora.  [#1286](https://github.com/theupdateframework/notary/pull/1286/)
+ If a certificate is used as a delegation public key, notary no longer warns if the certificate has expired, since notary should be relying on the role expiry instead. [#1263](https://github.com/theupdateframework/notary/pull/1263)
+ An error is now returned when importing keys if there were invalid PEM blocks. [#1260](https://github.com/theupdateframework/notary/pull/1260)
+ Notary server authentication credentials can now be provided as an environment variable `NOTARY_AUTH`, which should contain a base64-encoded "username:password" value.  [#1246](https://github.com/theupdateframework/notary/pull/1246)
+ Changefeeds are now supported for RethinkDB as well as SQL servers.  [#1214](https://github.com/theupdateframework/notary/pull/1214)
+ Notary CLI will now time out after 30 seconds if a username and password are not provided when authenticating to anotary server, fixing an issue where scripts for the notary CLI may hang forever. [#1200](https://github.com/theupdateframework/notary/pull/1200)
+ Fixed potential race condition in the signer keystore.  [#1198](https://github.com/theupdateframework/notary/pull/1198)
+ Notary now no longer provides the option to generate RSA keys for a repository, but externally generated RSA keys can still be imported as keys for a repository.  [#1191](https://github.com/theupdateframework/notary/pull/1191)
+ Fixed bug where the notary client would `ioutil.ReadAll` responses from the server without limiting the size.  [#1186](https://github.com/theupdateframework/notary/pull/1186)
+ Default notary CLI log level is now `warn`, and if the `-v` option is passed, it is at `info`.  [#1179](https://github.com/theupdateframework/notary/pull/1179)
+ Example Postgres config now includes an example of mutual TLS authentication between the server/signer and Postgres.  [#1160](https://github.com/theupdateframework/notary/pull/1160) [#1163](https://github.com/theupdateframework/notary/pull/1163/)
+ Fixed an error where piping the server authentication credentials via STDIN when scripting the notary CLI did not work.  [#1155](https://github.com/theupdateframework/notary/pull/1155)
+ If the server and signer configurations forget to specify `parseTime=true` when using MySQL, notary server and signer will automatically add the option.  [#1150](https://github.com/theupdateframework/notary/pull/1150)
+ Custom metadata can now be provided and read on a target when using the notary client as a library (not yet exposed on the CLI).  [#1146](https://github.com/theupdateframework/notary/pull/1146)
+ `notary init` now accepts a `--root-cert` and `--root-key` flag for use with privately generated certificates and keys.  [#1144](https://github.com/theupdateframework/notary/pull/1144)
+ `notary key generate` now accepts a `--role` flag as well as a `--output` flag.  This means it can generate new targets or delegation keys, and it can also output keys to a file instead of storing it in the default notary key store.  [#1134](https://github.com/theupdateframework/notary/pull/1134)
+ Newly generated keys are now stored encrypted and encoded in PKCS#8 format. **This is not forwards-compatible against notary<0.6.0 and docker<17.12.x. Also please note that docker>=17.12.x is not forwards compatible with notary<0.6.0.**. [#1130](https://github.com/theupdateframework/notary/pull/1130) [#1201](https://github.com/theupdateframework/notary/pull/1201)
+ Added support for wildcarded certificate IDs in the trustpinning configuration [#1126](https://github.com/theupdateframework/notary/pull/1126)
+ Added support using the client against notary servers which are hosted as subpath under another server (e.g. https://domain.com/notary instead of https://notary.com) [#1108](https://github.com/theupdateframework/notary/pull/1108)
+ If no changes were made to the targets file, you are no longer required to sign the target [#1104](https://github.com/theupdateframework/notary/pull/1104)
+ escrow placeholder [#1096](https://github.com/theupdateframework/notary/pull/1096)
+ Added support for wildcard suffixes for root certificates CNs for root keys, so that a single root certificate would be valid for multiple repositories [#1088](https://github.com/theupdateframework/notary/pull/1088)
+ Root key rotations now do not require all previous root keys sign new root metadata.  [#942](https://github.com/theupdateframework/notary/pull/942).
    + New keys are trusted if the root metadata file specifying the new key was signed by the previous root key/threshold
    + Root metadata can now be requested by version from the server, allowing clients with older root metadata to validate each new version one by one up to the current metadata
+ `notary key rotate` now accepts a flag specifying which key to rotate to [#942](https://github.com/theupdateframework/notary/pull/942)
+ Refactoring of the client to make it easier to use as a library and to inject dependencies:
    + References to GUN have now been changed to "imagename". [#1081](https://github.com/theupdateframework/notary/pull/1081)
    + `NewNotaryRepository` can now be provided with a remote store and changelist, as opposed to always constructing its own. [#1094](https://github.com/theupdateframework/notary/pull/1094)
    + If needed, the notary repository will be initialized first when publishing. [#1105](https://github.com/theupdateframework/notary/pull/1105)
    + `NewNotaryReository` now requires a non-nil cache store.  [#1185](https://github.com/theupdateframework/notary/pull/1185)
    + The "No valid trust data" error is now typed.  [#1212](https://github.com/theupdateframework/notary/pull/1212)
    + `TUFClient` was previously mistakenly exported, and is now unexported.  [#1215](https://github.com/theupdateframework/notary/pull/1215)
    + The notary client now has a `Repository` interface type to standardize `client.NotaryRepository`.  [#1220](https://github.com/theupdateframework/notary/pull/1220)
    + The constructor functions `NewFileCachedNotaryRepository` and `NewNotaryRepository` have been renamed, respectively, to `NewFileCachedRepository` and `NewRepository` to reduce redundancy. [#1226](https://github.com/theupdateframework/notary/pull/1226)
    + `NewRepository` returns an interface as opposed to the concrete type `NotaryRepository` it previously did.  `NotaryRepository` is also now an unexported concrete type. [#1226](https://github.com/theupdateframework/notary/pull/1226)
    + Key import/export logic has been moved from the `utils` package to the `trustmanager` package.  [#1250](https://github.com/theupdateframework/notary/pull/1250)


## [v0.5.0](https://github.com/docker/notary/releases/tag/v0.5.0) 11/14/2016
+ Non-certificate public keys in PEM format can now be added to delegation roles [#965](https://github.com/docker/notary/pull/965)
+ PostgreSQL support as a storage backend for Server and Signer [#920](https://github.com/docker/notary/pull/920)
+ Notary server's health check now fails if it cannot connect to the signer, since no new repositories can be created and existing repositories cannot be updated if the server cannot reach the signer [#952](https://github.com/docker/notary/pull/952)
+ Server runs its connectivity healthcheck to the server once every 10 seconds instead of once every minute. [#902](https://github.com/docker/notary/pull/902)
+ The keys on disk are now stored in the `~/.notary/private` directory, rather than in a key hierarchy that separates them by GUN and by role.  Notary will automatically migrate old-style directory layouts to the new style.  **This is not forwards-compatible against notary<0.4.2 and docker<=1.12** [#872](https://github.com/docker/notary/pull/872)
+ A new changefeed API has been added to Notary Server. It is only supported when using one of the relational database backends: MySQL, PostgreSQL, or SQLite.[#1019](https://github.com/docker/notary/pull/1019)

## [v0.4.3](https://github.com/docker/notary/releases/tag/v0.4.3) 1/3/2017
+ Fix build tags for static notary client binaries in linux [#1039](https://github.com/docker/notary/pull/1039)
+ Fix key import for exported delegation keys [#1067](https://github.com/docker/notary/pull/1067)

## [v0.4.2](https://github.com/docker/notary/releases/tag/v0.4.2) 9/30/2016
+ Bump the cross compiler to golang 1.7.1, since [1.6.3 builds binaries that could have non-deterministic bugs in OS X Sierra](https://groups.google.com/forum/#!msg/golang-dev/Jho5sBHZgAg/cq6d97S1AwAJ) [#984](https://github.com/docker/notary/pull/984)

## [v0.4.1](https://github.com/docker/notary/releases/tag/v0.4.1) 9/27/2016
+ Preliminary Windows support for notary client [#970](https://github.com/docker/notary/pull/970)
+ Output message to CLI when repo changes have been successfully published [#974](https://github.com/docker/notary/pull/974)
+ Improved error messages for client authentication errors and for the witness command [#972](https://github.com/docker/notary/pull/972)
+ Support for finding keys that are anywhere in the notary directory's "private" directory, not just under "private/root_keys" or "private/tuf_keys" [#981](https://github.com/docker/notary/pull/981)
+ Previously, on any error updating, the client would fall back on the cache.  Now we only do so if there is a network error or if the server is unavailable or missing the TUF data. Invalid TUF data will cause the update to fail - for example if there was an invalid root rotation. [#884](https://github.com/docker/notary/pull/884) [#982](https://github.com/docker/notary/pull/982)

## [v0.4.0](https://github.com/docker/notary/releases/tag/v0.4.0) 9/21/2016
+ Server-managed key rotations [#889](https://github.com/docker/notary/pull/889)
+ Remove `timestamp_keys` table, which stored redundant information [#889](https://github.com/docker/notary/pull/889)
+ Introduce `notary delete` command to delete local and/or remote repo data [#895](https://github.com/docker/notary/pull/895)
+ Introduce `notary witness` command to stage signatures for specified roles [#875](https://github.com/docker/notary/pull/875)
+ Add `-p` flag to offline commands to attempt auto-publish [#886](https://github.com/docker/notary/pull/886) [#912](https://github.com/docker/notary/pull/912) [#923](https://github.com/docker/notary/pull/923)
+ Introduce `notary reset` command to manage staged changes [#959](https://github.com/docker/notary/pull/959) [#856](https://github.com/docker/notary/pull/856)
+ Add `--rootkey` flag to `notary init` to provide a private root key for a repo [#801](https://github.com/docker/notary/pull/801)
+ Introduce `notary delegation purge` command to remove a specified key from all delegations [#855](https://github.com/docker/notary/pull/855)
+ Removed HTTP endpoint from notary-signer [#870](https://github.com/docker/notary/pull/870)
+ Refactored and unified key storage [#825](https://github.com/docker/notary/pull/825)
+ Batched key import and export now operate on PEM files (potentially with multiple blocks) instead of ZIP [#825](https://github.com/docker/notary/pull/825) [#882](https://github.com/docker/notary/pull/882)
+ Add full database integration test-suite [#824](https://github.com/docker/notary/pull/824) [#854](https://github.com/docker/notary/pull/854) [#863](https://github.com/docker/notary/pull/863)
+ Improve notary-server, trust pinning, and yubikey logging [#798](https://github.com/docker/notary/pull/798) [#858](https://github.com/docker/notary/pull/858) [#891](https://github.com/docker/notary/pull/891)
+ Warn if certificates for root or delegations are near expiry [#802](https://github.com/docker/notary/pull/802)
+ Warn if role metadata is near expiry [#786](https://github.com/docker/notary/pull/786)
+ Reformat CLI table output to use the `text/tabwriter` package [#809](https://github.com/docker/notary/pull/809)
+ Fix passphrase retrieval attempt counting and terminal detection [#906](https://github.com/docker/notary/pull/906)
+ Fix listing nested delegations [#864](https://github.com/docker/notary/pull/864)
+ Bump go version to 1.6.3, fix go1.7 compatibility [#851](https://github.com/docker/notary/pull/851) [#793](https://github.com/docker/notary/pull/793)
+ Convert docker-compose files to v2 format [#755](https://github.com/docker/notary/pull/755)
+ Validate root rotations against trust pinning [#800](https://github.com/docker/notary/pull/800)
+ Update fixture certificates for two-year expiry window [#951](https://github.com/docker/notary/pull/951)

## [v0.3.0](https://github.com/docker/notary/releases/tag/v0.3.0) 5/11/2016
+ Root rotations
+ RethinkDB support as a storage backend for Server and Signer
+ A new TUF repo builder that merges server and client validation
+ Trust Pinning: configure known good key IDs and CAs to replace TOFU.
+ Add --input, --output, and --quiet flags to notary verify command
+ Remove local certificate store. It was redundant as all certs were also stored in the cached root.json
+ Cleanup of dead code in client side key storage logic
+ Update project to Go 1.6.1
+ Reorganize vendoring to meet Go 1.6+ standard. Still using Godeps to manage vendored packages
+ Add targets by hash, no longer necessary to have the original target data available
+ Active Key ID verification during signature verification
+ Switch all testing from assert to require, reduces noise in test runs
+ Use alpine based images for smaller downloads and faster setup times
+ Clean up out of data signatures when re-signing content
+ Set cache control headers on HTTP responses from Notary Server
+ Add sha512 support for targets
+ Add environment variable for delegation key passphrase
+ Reduce permissions requested by client from token server
+ Update formatting for delegation list output
+ Move SQLite dependency to tests only so it doesn't get built into official images
+ Fixed asking for password to list private repositories
+ Enable using notary client with username/password in a scripted fashion
+ Fix static compilation of client
+ Enforce TUF version to be >= 1, previously 0 was acceptable although unused
+ json.RawMessage should always be used as *json.RawMessage due to concepts of addressability in Go and effects on encoding

## [v0.2](https://github.com/docker/notary/releases/tag/v0.2.0) 2/24/2016
+ Add support for delegation roles in `notary` server and client
+ Add `notary CLI` commands for managing delegation roles: `notary delegation`
    + `add`, `list` and `remove` subcommands
+ Enhance `notary CLI` commands for adding targets to delegation roles
    + `notary add --roles` and `notary remove --roles` to manipulate targets for delegations
+ Support for rotating the snapshot key to one managed by the `notary` server
+ Add consistent download functionality to download metadata and content by checksum
+ Update `docker-compose` configuration to use official mariadb image
    + deprecate `notarymysql`
    + default to using a volume for `data` directory
    + use separate databases for `notary-server` and `notary-signer` with separate users
+ Add `notary CLI` command for changing private key passphrases: `notary key passwd`
+ Enhance `notary CLI` commands for importing and exporting keys
+ Change default `notary CLI` log level to fatal, introduce new verbose (error-level) and debug-level settings
+ Store roles as PEM headers in private keys, incompatible with previous notary v0.1 key format
    + No longer store keys as `<KEY_ID>_role.key`, instead store as `<KEY_ID>.key`; new private keys from new notary clients will crash old notary clients
+ Support logging as JSON format on server and signer
+ Support mutual TLS between notary client and notary server

## [v0.1](https://github.com/docker/notary/releases/tag/v0.1) 11/15/2015
+ Initial non-alpha `notary` version
+ Implement TUF (the update framework) with support for root, targets, snapshot, and timestamp roles
+ Add PKCS11 interface to store and sign with keys in HSMs (i.e. Yubikey)
