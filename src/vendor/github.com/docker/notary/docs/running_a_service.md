<!--[metadata]>
+++
title = "Running a Notary service"
description = "Run your own notary service to host arbitrary content signing."
keywords = ["docker, notary, notary-server, notary server, notary-signer, notary signer"]
[menu.main]
parent="mn_notary"
weight=4
+++
<![end-metadata]-->

# Run a Notary service

This document is for anyone who wants to run their own Notary
service (such as those who want to use Notary with a
private Docker registry). Running a Notary service requires that you are already
familiar with using [Docker Engine](https://docs.docker.com/engine/userguide/)
and [Docker Compose](https://docs.docker.com/compose/overview/).

## Run a service for testing or development

The quickest way to spin up a full Notary service for testing and development
purposes is to use the Docker compose file in the
<a href="https://github.com/docker/notary" target="_blank">Notary project</a>.

```plain
$ git clone https://github.com/docker/notary.git
$ cd notary
$ docker-compose up
```

This will build the development Notary server and Notary signer images, and
start up containers for the Notary server, Notary signer, and the MySQL
database that both of them share.  The MySQL data is stored in a volume.

Notary server and Notary signer communicate over mutually authenticated TLS
(using the self-signed testing certs in the repository), and Notary server
listens for HTTPS traffic on port 4443.

By default, this development Notary server container runs with the testing
self-signed TLS certificates. In order to be able to successfully connect to
it, you will have to use the root CA file in `fixtures/root-ca.crt`.

For example, to connect using OpenSSL:

```bash
$ openssl s_client -connect <docker host>:4443 -CAfile fixtures/root-ca.crt -no_ssl3 -no_ssl2
```

To connect using the Notary Client CLI, please see [Getting Started](getting_started.md)
documentation. Please note that the version of Notary server and signer
should be greater than or equal to that of the Notary Client CLI to ensure feature compatibility,
i.e. if you are using Notary Client CLI 0.2, ensure you are using a server and signer tagged with
an equal or higher version than 0.2 from the <a href="https://github.com/docker/notary/releases">releases page</a>.

The self-signed certificate's subject name and subject alternative names are
`notary-server`, `notaryserver`, and `localhost`, so if your Docker host is not
on localhost (for example if you are using Docker Machine), you'll need to
update your hosts file such that the name `notary-server` is associated with
the IP address of your Docker host.

## Advanced configuration options

Both the Notary server and the Notary signer take
[JSON configuration files](reference/index.md). Pre-built images, such as
the [development images above](#run-a-service-for-testing-or-development)
provide these configuration files for you with some sane defaults.

However, for running in production, or if you just want to change those defaults
on your development service, you probably want to change those defaults.

### Running with different command line arguments

You can override the `docker run` command for the image if you want to pass
different command line options.  Both Notary server and Notary signer take
the following command line arguments:

- `-config=<config file>` - specify the path to the JSON configuration file.

- `-debug` - Passing this flag enables the debugging server on `localhost:8080`.
	The debugging server provides <a href="https://golang.org/pkg/net/http/pprof/" target="_blank">pprof</a>
	and <a href="https://golang.org/pkg/expvar/" target="_blank">expvar</a> endpoints.
	(Remember, this is localhost with respect to the running container - this endpoint is not
	exposed from the container).

	This option can also be set in the configuration file.

- `-logf=<format>` - This flag sets the output format for the logs. Possible
    formats are "json" and "logfmt".

    This option cannot be set in the configuration file, since some log
    messages are produced on startup before the configuration file has been
    read.

### Specifying your own configuration files

You can run the images with your own configuration files entirely.
You just need to mount your configuration directory, and then pass the
path to that configuration file as an argument to the `docker run` command.

### Overriding configuration file parameters using environment variables

You can also override the parameters of the configuration by
setting environment variables of the form `NOTARY_SERVER_<var>` or
`NOTARY_SIGNER_<var>`.

`var` is the ALL-CAPS, `"_"`-delimited path of keys from the top level of the
configuration JSON.

For instance, if you wanted to override the storage URL of the Notary server
configuration:

```json
"storage": {
  "backend": "mysql",
  "db_url": "dockercondemo:dockercondemo@tcp(notary-mysql)/dockercondemo"
}
```

You would need to set the environment variable `NOTARY_SERVER_STORAGE_DB_URL`,
because the `db_url` is in the `storage` section of the Notary server
configuration JSON.

Note that you cannot override a key whose value is another map.
For instance, setting
`NOTARY_SERVER_STORAGE='{"storage": {"backend": "memory"}}'` will not
set in-memory storage.  It just fails to parse.  You can only override keys
whose values are strings or numbers.

For example, let's say that you wanted to run a single Notary server instance:

- with your own TLS cert and keys
- with a local, in-memory signer service rather than using Notary signer,
- using a local, in-memory TUF metadata store rather than using MySQL
- produce JSON-formatted logs

One way to do this would be:

1. Generate your own TLS certificate and key as `server.crt` and `server.key`,
	and put them in the directory `/tmp/server-configdir`.

2. Write the following configuration file to `/tmp/server-configdir/config.json`:

		{
		  "server": {
		    "http_addr": ":4443",
		    "tls_key_file": "./server.key",
			"tls_cert_file": "./server.crt"
		  },
		  "trust_service": {
		    "type": "remote",
		    "hostname": "notarysigner",
		    "port": "7899",
		    "tls_ca_file": "./root-ca.crt",
		    "key_algorithm": "ecdsa",
		    "tls_client_cert": "./notary-server.crt",
		    "tls_client_key": "./notary-server.key"
		  },
		  "storage": {
		    "backend": "mysql",
		    "db_url": "server@tcp(mysql:3306)/notaryserver?parseTime=True"
		  }
		}

	Note that we are including a remote trust service and a database storage
	type in order to demonstrate how environment variables can override
	configuration parameters.

3. Run the following command (assuming you've already built or pulled a Notary server docker image):

		$ docker run \
			-p "4443:4443" \
			-v /tmp/server-configdir:/etc/docker/notary-server/ \
			-e NOTARY_SERVER_TRUST_SERVICE_TYPE=local \
			-e NOTARY_SERVER_STORAGE_BACKEND=memory \
			-e NOTARY_SERVER_LOGGING_LEVEL=debug \
			notary_server \
				-config=/etc/docker/notary-server/config.json \
				-logf=json
		{"level":"info","msg":"Version: 0.2, Git commit: 619f8cf","time":"2016-02-25T00:53:59Z"}
		{"level":"info","msg":"Using local signing service, which requires ED25519. Ignoring all other trust_service parameters, including keyAlgorithm","time":"2016-02-25T00:53:59Z"}
		{"level":"info","msg":"Using memory backend","time":"2016-02-25T00:53:59Z"}
		{"level":"info","msg":"Starting Server","time":"2016-02-25T00:53:59Z"}
		{"level":"info","msg":"Enabling TLS","time":"2016-02-25T00:53:59Z"}
		{"level":"info","msg":"Starting on :4443","time":"2016-02-25T00:53:59Z"}

You can do the same using [Docker
Compose](https://docs.docker.com/compose/overview/) by setting volumes,
environment variables, and overriding the default command for the Notary server
containers in the Compose file.


## Recommendations for deploying in production

When moving from development to production there are a number of considerations
that must be made to ensure security and scalability.

### Certificates

The Notary repository includes sample certificates in the fixtures directory.
When you initialize a development service using the provided docker-compose.yml
file, these sample certificates are used to create a more production like
environment.

**You must acquire _your_ _own_ _certificates_ to use in a production deployment.**

The sample private key files in the Notary repository are obviously public knowledge
and using them in a production deployment is highly insecure.

### Databases

The server and signer each require a database. These should be separate databases
with different users. The users should be limited in their permissions. We recommend
giving the following MySQL (or equivalent) permissions to the users restricted to
only their own databases:

- Notary server database user: `SELECT, INSERT, UPDATE, DELETE`
- Notary signer database user: `SELECT, INSERT, UPDATE, DELETE`

### High Availability

Most production users will want to increase availability by running multiple instances
of both the server and signer applications. These can be scaled arbitrarily and
independently. The database may also be scaled independently but this is left as
and exercise for experienced DBAs and Operations teams. A typical deployment will
look like the below diagram:

![Notary server Deployment Diagram](https://cdn.rawgit.com/docker/notary/09f81717080f53276e6881ece57cbbbf91b8e2a7/docs/images/service-deployment.svg)

In the diagram, a load balancer routes external traffic to a cluster of Notary server
instances. These may make requests to Notary signer instances if either a) signing
is required, or b) key generation is required. The requests from a Notary server
to a Notary signer cluster are router via an internal load balancer.

Notary can be used with a CDN or other caching system. All GET requests for JSON
files may be cached indefinitely __except__ URLs matching:

- `*/root.json`
- `*/timestamp.json`

All other requests for JSON files include sha256 checksums of the file being requested
and are therefore immutable. Requests for JSON files make up the vast majority of
all notary requests. Requests for anything other than a GET of a JSON file should
not be cached.

## Related information

* [Notary service architecture](service_architecture.md)
* [Notary configuration files](reference/index.md)
