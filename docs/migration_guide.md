# Harbor upgrade and database migration guide

When upgrading your existing Habor instance to a newer version, you may need to migrate the data in your database. Refer to [change log](../migration/changelog.md) to find out whether there is any change in the database. If there is, you should go through the database migration process. Since the migration may alter the database schema, you should **always** back up your data before any migration. 

*If your install Harbor for the first time, or the database version is the same as that of the lastest version, you do not need any database migration.*


**NOTE:** You must backup your data before any data migration.

### Upgrading Harbor and migrating data

1. Log in to the machine that Harbor runs on, stop and remove existing Harbor service if it is still running:

    ``` 
    cd make/
    docker-compose down
    ```

2.  Back up Harbor's current source code so that you can roll back to the current version when it is necessary.
    ```sh
    cd ../..
    mv harbor /tmp/harbor
    ```

3. Get the lastest source code from Github:
    ```sh
    git clone https://github.com/vmware/harbor
    ```
 
4. Before upgrading Harbor, perform database migration first.
The directory **migration/** contains the tool for migration. The first step is to update values of `db_username`, `db_password`, `db_port`, `db_name` in **migration.cfg** so that they match your system's configuration. 

5. The migration tool is delivered as a container, so you should build the image from its Dockerfile:
    ```
    cd migration/
    
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

8. Change to `make/` directory, configure Harbor by modifying the file `harbor.cfg`, you may need to refer to the configuration files you've backed up during step 2. Refer to [Installation & Configuration Guide ](../docs/installation_guide.md) for more info.

9. If HTTPS has been enabled for Harbor before, restore the `nginx.conf` and key/certificate files from the backup files in Step 2. Refer to [Configuring Harbor with HTTPS Access](../docs/configure_https.md) for more info.

10. Under the directory `make/`, run the `./prepare` script to generate necessary config files.
 
11. Rebuild Harbor and restart the registry service

    ```
    docker-compose up --build -d
    ```

### Roll back from an upgrade
For any reason, if you want to roll back to the previous version of Harbor, follow the below steps:

1. Stop and remove the current Harbor service if it is still running.

    ``` 
    cd make/
    docker-compose down
    ```
2. Restore database from backup file in `/path/to/backup` .

    ```
    docker run -ti --rm -v /data/database:/var/lib/mysql -v /path/to/backup:/harbor-migration/backup migrate-tool restore
    ```

3. Remove current source code of Harbor.
    ``` 
    rm -rf harbor
    ```

4. Restore the source code of an older version of Harbor. 
    ```sh
    mv /tmp/harbor harbor
    ```

5. Restart Harbor service using the previous configuration.
    ```sh
    cd make/
    docker-compose up --build -d
    ```
    
### Migration tool reference
- Use `help` command to show instructions of the migration tool:

    ```docker run --rm migrate-tool help```
    
- Use `test` command to test mysql connection:

    ```docker run --rm -v /data/database:/var/lib/mysql migrate-tool test```

