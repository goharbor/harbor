# Google Cloud Spanner

## Usage

The DSN must be given in the following format.

`spanner://projects/{projectId}/instances/{instanceId}/databases/{databaseName}`

See [Google Spanner Documentation](https://cloud.google.com/spanner/docs) for details.


| Param | WithInstance Config | Description |
| ----- | ------------------- | ----------- |
| `x-migrations-table` | `MigrationsTable` | Name of the migrations table |
| `url` | `DatabaseName` | The full path to the Spanner database resource. If provided as part of `Config` it must not contain a scheme or query string to match the format `projects/{projectId}/instances/{instanceId}/databases/{databaseName}`|
| `projectId` || The Google Cloud Platform project id
| `instanceId` || The id of the instance running Spanner
| `databaseName` || The name of the Spanner database


> **Note:** Google Cloud Spanner migrations can take a considerable amount of 
> time. The migrations provided as part of the example take about 6 minutes to 
> run on a small instance.
>
> ```log
> 1481574547/u create_users_table (21.354507597s)
> 1496539702/u add_city_to_users (41.647359754s)
> 1496601752/u add_index_on_user_emails (2m12.155787369s)
> 1496602638/u create_books_table (2m30.77299181s)

## Testing

To unit test the `spanner` driver, `SPANNER_DATABASE` needs to be set. You'll
need to sign-up to Google Cloud Platform (GCP) and have a running Spanner
instance since it is not possible to run Google Spanner outside GCP.