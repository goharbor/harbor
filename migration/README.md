# Harbor upgrade and database migration guide

When upgrading your existing Habor instance to a newer version, you may need to migrate the data in your database. Refer to [change log](changelog.md) to find out whether there is any change in the database. If there is, you should go through the database migration process. Since the migration may alter the database schema, you should **always** back up your data before any migration. 

*If your install Harbor for the first time, or the database version is the same as that of the lastest version, you do not need any database migration.*


**NOTE:** You must backup your data before any data migration.

### Upgrading Harbor and migrating data

1. Log in to the machine that Harbor runs on, back up Harbor's configuration files. 
    ```sh
    mkdir -p /tmp/harbor/config
    cp -r Deploy/config /tmp/harbor/config
    cp Deploy/harbor.cfg /tmp/harbor
    ```

2. Next, stop existing Harbor service if it is still running:

    ``` 
    cd ../../Deploy/
    docker-compose down
    ```

3. Get the lastest source code from Github:
    ```sh
    $ git clone https://github.com/vmware/harbor
    ```
 
4. Before upgrading Harbor, perform data migration first.
The directory **migration/** contains the tool for migration. The first step is to update values of `db_username`, `db_password`, `db_port`, `db_name` in **migration.cfg** so that they match your system's configuration. 

5. The migration tool is delivered as a container, so you should build the image from its Dockerfile:
    ```
    cd migration/harbor-migration
    
    docker build -t migrate-tool .
    ```

6. Back up database to a directory such as `/path/to/backup`. You need to create the directory if it does not exist. 

    ```
    docker run -ti --rm -v /data/database:/var/lib/mysql -v /path/to/backup:/harbor-migration/backup migrate-tool backup
    ```

7.  Upgrade database schema and migrate data:

    ```
    docker run -ti --rm -v /data/database:/var/lib/mysql migrate-tool up head
    ```

8. Change to `Deploy/` directory, configure Harbor by updating the file `harbor.cfg`, you may need to refer to the configuration files you backed up during step 1. Refer to [Installation & Configuration Guide ](../docs/installation_guide.md) for more info.

9. If HTTPS has been enabled for Harbor before, restore the `nginx.conf` and key/certificate files from the backup files in Step 1. Refer to [Configuring Harbor with HTTPS Access](../docs/configure_https.md) for more info.

10. Run the `./prepare` script to generate necessary config files.
 
11. Rebuild Harbor and restart the registry service

    ```
    docker-compose up --build -d
    ```

### Migration tool reference
- Use `help` command to show instructions of the migration tool:

    ```docker run --rm migrate-tool help```
    
- Use `test` command to test mysql connection:

    ```docker run --rm -v /data/database:/var/lib/mysql migrate-tool test```

- Restore database from backup file in `/path/to/backup`

    ```
    docker run -ti --rm -v /data/database:/var/lib/mysql -v /path/to/backup:/harbor-migration/backup migrate-tool restore
    ```
