# Cassandra

* Drop command will not work on Cassandra 2.X because it rely on
system_schema table which comes with 3.X
* Other commands should work properly but are **not tested**


## Usage
`cassandra://host:port/keyspace?param1=value&param2=value2`


| URL Query  | Default value | Description |
|------------|-------------|-----------|
| `x-migrations-table` | schema_migrations | Name of the migrations table |
| `port` | 9042 | The port to bind to  |
| `consistency` | ALL | Migration consistency
| `protocol` |  | Cassandra protocol version (3 or 4)
| `timeout` | 1 minute | Migration timeout
| `username` | nil | Username to use when authenticating. |
| `password` | nil | Password to use when authenticating. |


`timeout` is parsed using [time.ParseDuration(s string)](https://golang.org/pkg/time/#ParseDuration)


## Upgrading from v1

1. Write down the current migration version from schema_migrations
2. `DROP TABLE schema_migrations`
4. Download and install the latest migrate version.
5. Force the current migration version with `migrate force <current_version>`.
