# go-tuf

[![build](https://github.com/theupdateframework/go-tuf/workflows/build/badge.svg)](https://github.com/theupdateframework/go-tuf/actions?query=workflow%3Abuild) [![Coverage Status](https://coveralls.io/repos/github/theupdateframework/go-tuf/badge.svg)](https://coveralls.io/github/theupdateframework/go-tuf) [![PkgGoDev](https://pkg.go.dev/badge/github.com/theupdateframework/go-tuf)](https://pkg.go.dev/github.com/theupdateframework/go-tuf) [![Go Report Card](https://goreportcard.com/badge/github.com/theupdateframework/go-tuf)](https://goreportcard.com/report/github.com/theupdateframework/go-tuf)

This is a Go implementation of [The Update Framework (TUF)](http://theupdateframework.com/),
a framework for securing software update systems.

## Directory layout

A TUF repository has the following directory layout:

```bash
.
├── keys
├── repository
│   └── targets
└── staged
    └── targets
```

The directories contain the following files:

- `keys/` - signing keys (optionally encrypted) with filename pattern `ROLE.json`
- `repository/` - signed metadata files
- `repository/targets/` - hashed target files
- `staged/` - either signed, unsigned or partially signed metadata files
- `staged/targets/` - unhashed target files

## CLI

`go-tuf` provides a CLI for managing a local TUF repository.

### Install

`go-tuf` is tested on Go versions 1.18.

```bash
go get github.com/theupdateframework/go-tuf/cmd/tuf
```

### Commands

#### `tuf init [--consistent-snapshot=false]`

Initializes a new repository.

This is only required if the repository should not generate consistent
snapshots (i.e. by passing `--consistent-snapshot=false`). If consistent
snapshots should be generated, the repository will be implicitly
initialized to do so when generating keys.

#### `tuf gen-key [--expires=<days>] <role>`

Prompts the user for an encryption passphrase (unless the
`--insecure-plaintext` flag is set), then generates a new signing key and
writes it to the relevant key file in the `keys` directory. It also stages
the addition of the new key to the `root` metadata file. Alternatively, passphrases
can be set via environment variables in the form of `TUF_{{ROLE}}_PASSPHRASE`

#### `tuf revoke-key [--expires=<days>] <role> <id>`

Revoke a signing key

The key will be removed from the root metadata file, but the key will remain in the
"keys" directory if present.

#### `tuf add [<path>...]`

Hashes files in the `staged/targets` directory at the given path(s), then
updates and stages the `targets` metadata file. Specifying no paths hashes all
files in the `staged/targets` directory.

#### `tuf remove [<path>...]`

Stages the removal of files with the given path(s) from the `targets` metadata file
(they get removed from the filesystem when the change is committed). Specifying
no paths removes all files from the `targets` metadata file.

#### `tuf snapshot [--expires=<days>]`

Expects a staged, fully signed `targets` metadata file and stages an appropriate
`snapshot` metadata file. Optionally one can set number of days after which
the `snapshot` metadata will expire.

#### `tuf timestamp [--expires=<days>]`

Stages an appropriate `timestamp` metadata file. If a `snapshot` metadata file is staged,
it must be fully signed. Optionally one can set number of days after which
the timestamp metadata will expire.

#### `tuf sign <metadata>`

Signs the given role's staged metadata file with all keys present in the `keys`
directory for that role.

#### `tuf commit`

Verifies that all staged changes contain the correct information and are signed
to the correct threshold, then moves the staged files into the `repository`
directory. It also removes any target files which are not in the `targets`
metadata file.

#### `tuf regenerate [--consistent-snapshot=false]`

Note: Not supported yet

Recreates the `targets` metadata file based on the files in `repository/targets`.

#### `tuf clean`

Removes all staged metadata files and targets.

#### `tuf root-keys`

Outputs a JSON serialized array of root keys to STDOUT. The resulting JSON
should be distributed to clients for performing initial updates.

#### `tuf set-threshold <role> <threshold>`

Sets `role`'s threshold (required number of keys for signing) to
`threshold`.

#### `tuf get-threshold <role>`

Outputs `role`'s threshold (required number of keys for signing).

#### `tuf change-passphrase <role>`

Changes the passphrase for given role keys file. The CLI supports reading
both the existing and the new passphrase via the following environment
variables - `TUF_{{ROLE}}_PASSPHRASE` and respectively `TUF_NEW_{{ROLE}}_PASSPHRASE`

#### `tuf payload <metadata>`

Outputs the metadata file for a role in a ready-to-sign (canonicalized) format.

See also `tuf sign-payload` and `tuf add-signatures`.

#### `tuf sign-payload --role=<role> <path>`

Sign a file (outside of the TUF repo) using keys (in the TUF keys database,
typically produced by `tuf gen-key`) for the given `role` (from the TUF repo).

Typically, `path` will be a file containing the output of `tuf payload`.

See also `tuf add-signatures`.

#### `tuf add-signatures --signatures <sig_file> <metadata>`


Adds signatures (the output of `tuf sign-payload`) to the given role metadata file.

If the signature does not verify, it will not be added.

#### `tuf status --valid-at <date> <role>`

Check if the role's metadata will be expired on the given date. 

#### Usage of environment variables

The `tuf` CLI supports receiving passphrases via environment variables in
the form of `TUF_{{ROLE}}_PASSPHRASE` for existing ones and
`TUF_NEW_{{ROLE}}_PASSPHRASE` for setting new ones.

For a list of supported commands, run `tuf help` from the command line.

### Examples

The following are example workflows for managing a TUF repository with the CLI.

The `tree` commands do not need to be run, but their output serve as an
illustration of what files should exist after performing certain commands.

Although only two machines are referenced (i.e. the "root" and "repo" boxes),
the workflows can be trivially extended to many signing machines by copying
staged changes and signing on each machine in turn before finally committing.

Some key IDs are truncated for illustrative purposes.

#### Create signed root metadata file

Generate a root key on the root box:

```bash
$ tuf gen-key root
Enter root keys passphrase:
Repeat root keys passphrase:
Generated root key with ID 184b133f

$ tree .
.
├── keys
│   └── root.json
├── repository
└── staged
    ├── root.json
    └── targets
```

Copy `staged/root.json` from the root box to the repo box and generate targets,
snapshot and timestamp keys:

```bash
$ tree .
.
├── keys
├── repository
└── staged
    ├── root.json
    └── targets

$ tuf gen-key targets
Enter targets keys passphrase:
Repeat targets keys passphrase:
Generated targets key with ID 8cf4810c

$ tuf gen-key snapshot
Enter snapshot keys passphrase:
Repeat snapshot keys passphrase:
Generated snapshot key with ID 3e070e53

$ tuf gen-key timestamp
Enter timestamp keys passphrase:
Repeat timestamp keys passphrase:
Generated timestamp key with ID a3768063

$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
└── staged
    ├── root.json
    └── targets
```

Copy `staged/root.json` from the repo box back to the root box and sign it:

```bash
$ tree .
.
├── keys
│   ├── root.json
├── repository
└── staged
    ├── root.json
    └── targets

$ tuf sign root.json
Enter root keys passphrase:
```

The staged `root.json` can now be copied back to the repo box ready to be
committed alongside other metadata files.

#### Alternate signing flow

Instead of manually copying `root.json` into the TUF repository on the root box,
you can use the `tuf payload`, `tuf sign-payload`, `tuf add-signatures` flow.

On the repo box, get the `root.json` payload in a canonical format:

``` bash
$ tuf payload root.json > root.json.payload
```

Copy `root.json.payload` to the root box and sign it:


``` bash
$ tuf sign-payload --role=root root.json.payload > root.json.sigs
Enter root keys passphrase:
```

Copy `root.json.sigs` back to the repo box and import the signatures:

``` bash
$ tuf add-signatures --signatures root.json.sigs root.json
```

This achieves the same state as the above flow for the repo box:

```bash
$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
└── staged
    ├── root.json
    └── targets
```

#### Add a target file

Assuming a staged, signed `root` metadata file and the file to add exists at
`staged/targets/foo/bar/baz.txt`:

```bash
$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
└── staged
    ├── root.json
    └── targets
        └── foo
            └── bar
                └── baz.txt

$ tuf add foo/bar/baz.txt
Enter targets keys passphrase:

$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
└── staged
    ├── root.json
    ├── targets
    │   └── foo
    │       └── bar
    │           └── baz.txt
    └── targets.json

$ tuf snapshot
Enter snapshot keys passphrase:

$ tuf timestamp
Enter timestamp keys passphrase:

$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
└── staged
    ├── root.json
    ├── snapshot.json
    ├── targets
    │   └── foo
    │       └── bar
    │           └── baz.txt
    ├── targets.json
    └── timestamp.json

$ tuf commit

$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
```

#### Remove a target file

Assuming the file to remove is at `repository/targets/foo/bar/baz.txt`:

```bash
$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged

$ tuf remove foo/bar/baz.txt
Enter targets keys passphrase:

$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
    └── targets.json

$ tuf snapshot
Enter snapshot keys passphrase:

$ tuf timestamp
Enter timestamp keys passphrase:

$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
    ├── snapshot.json
    ├── targets.json
    └── timestamp.json

$ tuf commit

$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
└── staged
```

#### Regenerate metadata files based on targets tree (Note: Not supported yet)

```bash
$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged

$ tuf regenerate
Enter targets keys passphrase:

$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
    └── targets.json

$ tuf snapshot
Enter snapshot keys passphrase:

$ tuf timestamp
Enter timestamp keys passphrase:

$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
    ├── snapshot.json
    ├── targets.json
    └── timestamp.json

$ tuf commit

$ tree .
.
├── keys
│   ├── snapshot.json
│   ├── targets.json
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
```

#### Update timestamp.json

```bash
$ tree .
.
├── keys
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged

$ tuf timestamp
Enter timestamp keys passphrase:

$ tree .
.
├── keys
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
    └── timestamp.json

$ tuf commit

$ tree .
.
├── keys
│   └── timestamp.json
├── repository
│   ├── root.json
│   ├── snapshot.json
│   ├── targets
│   │   └── foo
│   │       └── bar
│   │           └── baz.txt
│   ├── targets.json
│   └── timestamp.json
└── staged
```

#### Adding a new root key

Copy `staged/root.json` to the root box and generate a new root key on the root box:

```bash
$ tuf gen-key root
$ tuf sign root.json
```

Copy `staged/root.json` from the root box and commit:

```bash
$ tuf commit
```

#### Rotating root key(s)

Copy `staged/root.json` to the root box to do the rotation, where `abcd` is the keyid of the key that is being replaced:

```bash
$ tuf gen-key root
$ tuf revoke-key root abcd
$ tuf sign root.json
```

Note that `revoke-key` removes the old key from `root.json`, but the key remains in the `keys/` directory on the root box as it is needed to sign the next `root.json`. After this signing is done, the old key may be removed from `keys/`. Any number of keys may be added or revoked during this step, but ensure that at least a threshold of valid keys remain.

Copy `staged/root.json` from the root box to commit:

```bash
$ tuf commit
```

## Client

For the client package, see https://godoc.org/github.com/theupdateframework/go-tuf/client.

For the client CLI, see https://github.com/theupdateframework/go-tuf/tree/master/cmd/tuf-client.

## Contributing and Development

For local development, `go-tuf` requires Go version 1.18.

The [Python interoperability tests](client/python_interop/) require Python 3
(available as `python` on the `$PATH`) and the [`python-tuf`
package](https://github.com/theupdateframework/python-tuf) installed (`pip
install tuf`). To update the data for these tests requires Docker and make (see
test data [README.md](client/python_interop/testdata/README.md) for details).

Please see [CONTRIBUTING.md](docs/CONTRIBUTING.md) for contribution guidelines before making your first contribution!
