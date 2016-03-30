# MySQL Driver

### See [issue #1](https://github.com/mattes/migrate/issues/1#issuecomment-58728186) before using this driver!

* Runs migrations in transcations.
  That means that if a migration failes, it will be safely rolled back.
* Tries to return helpful error messages.
* Stores migration version details in table ``schema_migrations``.
  This table will be auto-generated.


## Usage

```bash
migrate -url mysql://user@tcp(host:port)/database -path ./db/migrations create add_field_to_table
migrate -url mysql://user@tcp(host:port)/database -path ./db/migrations up
migrate help # for more info
```

See full [DSN (Data Source Name) documentation](https://github.com/go-sql-driver/mysql/#dsn-data-source-name).

## Authors

* Matthias Kadenbach, https://github.com/mattes