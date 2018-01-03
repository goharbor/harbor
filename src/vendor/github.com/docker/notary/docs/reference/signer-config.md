<!--[metadata]>
+++
title = "Signer Configuration"
description = "Configuring the notary client, server and signer."
keywords = ["docker, notary, notary-client, notary-server, notary server, notary-signer, notary signer"]
[menu.main]
parent="mn_notary_config"
+++
<![end-metadata]-->


# Notary signer configuration file

This document is for those who are [running their own Notary service](../running_a_service.md) who
want to specify custom options.

## Overview

Notary signer [requires environment variables](#environment-variables-required-if-using-mysql)
to encrypt private keys at rest. It also requires a configuration file, the path to which is
specified on the command line using the `-config` flag.

Here is a full signer configuration file example; please click on the top level JSON keys to
learn more about the configuration section corresponding to that key:

<pre><code class="language-json">{
  <a href="#server-section-required">"server"</a>: {
    "grpc_addr": ":7899",
    "tls_cert_file": "./fixtures/notary-signer.crt",
    "tls_key_file": "./fixtures/notary-signer.key",
    "client_ca_file": "./fixtures/notary-server.crt"
  },
  <a href="../common-configs/#logging-section-optional">"logging"</a>: {
    "level": 2
  },
  <a href="#storage-section-required">"storage"</a>: {
    "backend": "mysql",
    "db_url": "user:pass@tcp(notarymysql:3306)/databasename?parseTime=true",
    "default_alias": "passwordalias1"
  },
  <a href="../common-configs/#reporting-section-optional">"reporting"</a>: {
    "bugsnag": {
      "api_key": "c9d60ae4c7e70c4b6c4ebd3e8056d2b8",
      "release_stage": "production"
    }
  }
}
</code></pre>

## server section (required)

"server" in this case refers to Notary signer's HTTP/GRPC server, not
"Notary server".

Example:

```json
"server": {
  "grpc_addr": ":7899",
  "tls_cert_file": "./fixtures/notary-signer.crt",
  "tls_key_file": "./fixtures/notary-signer.key",
  "client_ca_file": "./fixtures/notary-server.crt"
}
```

<table>
	<tr>
		<th>Parameter</th>
		<th>Required</th>
		<th>Description</th>
	</tr>
	<tr>
		<td valign="top"><code>grpc_addr</code></td>
		<td valign="top">yes</td>
		<td valign="top">The TCP address (IP and port) to listen for GRPC
			traffic.  Examples:
			<ul>
			<li><code>":7899"</code> means listen on port 7899 on all IPs (and
				hence all interfaces, such as those listed when you run
				<code>ifconfig</code>)</li>
			<li><code>"127.0.0.1:7899"</code> means listen on port 7899 on
				localhost only.  That means that the server will not be
				accessible except locally (via SSH tunnel, or just on a local
				terminal)</li>
			</ul>
		</td>
	</tr>
	<tr>
		<td valign="top"><code>tls_key_file</code></td>
		<td valign="top">yes</td>
		<td valign="top">The path to the private key to use for
			GRPC TLS. The path is relative to the directory of the
			configuration file.</td>
	</tr>
	<tr>
		<td valign="top"><code>tls_cert_file</code></td>
		<td valign="top">yes</td>
		<td valign="top">The path to the certificate to use for
			GRPC TLS. The path is relative to the directory of the
			configuration file.</td>
	</tr>
	<tr>
		<td valign="top"><code>client_ca_file</code></td>
		<td valign="top">no</td>
		<td valign="top">The root certificate to trust for
			mutual authentication. If provided, any clients connecting to
			Notary signer will have to have a client certificate signed by
			this root. If not provided, mutual authentication will not be
			required. The path is relative to the directory of the
			configuration file.</td>
	</tr>
</table>


## storage section (required)

This is used to store encrypted private keys.  We only support MySQL, PostgreSQL
or an in-memory store, currently.

Example:

```json
"storage": {
  "backend": "mysql",
  "db_url": "user:pass@tcp(notarymysql:3306)/databasename?parseTime=true",
  "default_alias": "passwordalias1"
}
```

<table>
	<tr>
		<th>Parameter</th>
		<th>Required</th>
		<th>Description</th>
	</tr>
	<tr>
		<td valign="top"><code>backend</code></td>
		<td valign="top">yes</td>
		<td valign="top">Must be <code>"mysql"</code>, <code>"postgres"</code> or <code>"memory"</code>.
			If <code>"memory"</code> is selected, the <code>db_url</code>
			is ignored.</td>
	</tr>
	<tr>
		<td valign="top"><code>db_url</code></td>
		<td valign="top">yes if not <code>memory</code></td>
		<td valign="top">The <a href="https://github.com/go-sql-driver/mysql">
			the Data Source Name used to access the DB.</a>
			(note: please include <code>parseTime=true</code> as part of the DSN)</td>
	</tr>
	<tr>
		<td valign="top"><code>default_alias</code></td>
		<td valign="top">yes if not <code>memory</code></td>
		<td valign="top">This parameter specifies the alias of the current
			password used to encrypt the private keys in the DB.  All new
			private keys will be encrypted using this password, which
			must also be provided as the environment variable
			<code>NOTARY_SIGNER_&lt;DEFAULT_ALIAS_VALUE&gt;</code>.
			Please see the <a href="#environment-variables-required-if-using-mysql">environment variable</a>
			section for more information.</td>
	</tr>
</table>


## Environment variables (required if using MySQL)

Notary signer stores the private keys in encrypted form.
The alias of the passphrase used to encrypt the keys is also stored.  In order
to encrypt the keys for storage and decrypt the keys for signing, the
passphrase must be passed in as an environment variable.

For example, the configuration above specifies the default password alias to be
`passwordalias1`.

If this configuration is used, then you must:

`export NOTARY_SIGNER_PASSWORDALIAS1=mypassword`

so that the Notary signer knows to encrypt all keys with the passphrase
`mypassword`, and to decrypt any private key stored with password alias
`passwordalias1` with the passphrase `mypassword`.

Older passwords may also be provided as environment variables.

Let's say that you wanted to change the password that is used to create new
keys (rotating the passphrase and re-encrypting all the private keys is not
supported yet).

You could change the config to look like:

```json
"storage": {
  "backend": "mysql",
  "db_url": "user:pass@tcp(notarymysql:3306)/databasename?parseTime=true",
  "default_alias": "passwordalias2"
}
```

Then you can set:

```bash
export NOTARY_SIGNER_PASSWORDALIAS1=mypassword
export NOTARY_SIGNER_PASSWORDALIAS2=mynewfancypassword
```

That way, all new keys will be encrypted and decrypted using the passphrase
`mynewfancypassword`, but old keys that were encrypted using the passphrase
`mypassword` can still be decrypted.

The environment variables for the older passwords are optional, but Notary
Signer will not be able to decrypt older keys if they are not provided, and
attempts to sign data using those keys will fail.

## Hot logging level reload
We don't support completely reloading notary signer configuration files yet at present. What we support for Linux and OSX now is:
- increase logging level by signaling `SIGUSR1`
- decrease logging level by signaling `SIGUSR2`

No signals and no dynamic logging level changes are supported for Windows yet.

Example:

To increase logging level
```
$ kill -s SIGUSR1 PID

or

$ docker exec -i CONTAINER_ID kill -s SIGUSR1 PID
```

To decrease logging level
```
$ kill -s SIGUSR2 PID

or

$ docker exec -i CONTAINER_ID kill -s SIGUSR2 PID
```
PID is the process id of `notary-signer` and it may not the PID 1 process if you are running
the container with some kind of wrapper startup script or something.

You can get the PID of `notary-signer` through
```
$ docker exec CONTAINER_ID ps aux

or

$ ps aux | grep "notary-signer -config" | grep -v "grep"
```


## Related information

* [Notary Server Configuration File](server-config.md)
* [Configuration sections common to the Notary server and signer](common-configs.md)
