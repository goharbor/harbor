# Harbor upgrade and database migration guide

When upgrading your existing Habor instance to a newer version, you may need to migrate the data in your database. Refer to [change log](../tools/migration/changelog.md) to find out whether there is any change in the database. If there is, you should go through the database migration process. Since the migration may alter the database schema, you should **always** back up your data before any migration.

*If your install Harbor for the first time, or the database version is the same as that of the lastest version, you do not need any database migration.*

**NOTE:** 
- From v1.2 on, you need to use the release version as the tag of the migrator image. 'latest' is no longer used for new release.

- You must back up your data before any data migration.

- To migrate harbor OVA, please refer [migrate OVA guide](migrate_ova_guide.md)

### Upgrading Harbor and migrating data

1. Log in to the host that Harbor runs on, stop and remove existing Harbor instance if it is still running:

    ```
    cd harbor
    docker-compose down
    ```

2.  Back up Harbor's current files so that you can roll back to the current version when it is necessary.
    ```sh
    cd ..
    mv harbor /my_backup_dir/harbor
    ```

3. Get the lastest Harbor release package from Github:
   https://github.com/vmware/harbor/releases

4. Before upgrading Harbor, perform database migration first.  The migration tool is delivered as a docker image, so you should pull the image from docker hub. Replace [tag] with the release version of Harbor (e.g. 1.2) in the below command:

    ```
    docker pull vmware/harbor-db-migrator:[tag]
    ```

5. Back up database to a directory such as `/path/to/backup`. You need to create the directory if it does not exist.  Also, note that the username and password to access the db are provided via environment variable "DB_USR" and "DB_PWD". 

    **NOTE:** Upgrade from harbor 1.2 or older to harbor 1.3 must use `vmware/migratorharbor-db-migrator:1.2`. Because DB engine replaced by MariaDB in harbor 1.3

    ```
    docker run -ti --rm -e DB_USR=root -e DB_PWD=xxxx -v /data/database:/var/lib/mysql -v /path/to/backup:/harbor-migration/backup vmware/harbor-db-migrator:[tag] backup
    ```

6.  Upgrade database schema and migrate data.
Please note that you need to use the migrator of the targeted version, as an example, to upgrade to v1.6.0, you need to use the migrator with version v1.6.0. Migrators above v1.4 are available at goharbor/harbor-migrator in docker hub.

    ```
    docker run -ti --rm -e DB_USR=root -e DB_PWD=xxxx -v /data/database:/var/lib/mysql vmware/harbor-db-migrator:[tag] up
    ```

     **NOTE:** Some errors like

    ```
    [ERROR] Missing system table mysql.roles_mapping; please run mysql_upgrade to create it
    [ERROR] Incorrect definition of table mysql.event: expected column 'sql_mode' at position ... ...
    [ERROR] mysqld: Event Scheduler: An error occurred when initializing system tables. Disabling the Event Scheduler.
    [Warning] Failed to load slave replication state from table mysql.gtid_slave_pos: 1146: Table 'mysql.gtid_slave_pos' doesn't exist
    ```
    will be occurred during upgrading from harbor 1.2 to harbor 1.3, just ignore them if harbor can start successfully.

7. Unzip the new Harbor package and change to `./harbor` as the working directory. Configure Harbor by modifying the file `harbor.cfg`,

  - Configure Harbor by modifying the file `harbor.cfg`,
you may need to refer to the configuration files you've backed up during step 2.
Refer to [Installation & Configuration Guide ](../docs/installation_guide.md) for more information.
Since the content and format of `harbor.cfg` may have been changed in the new release, **DO NOT directly copy `harbor.cfg` from previous version of Harbor.**

	**IMPORTANT:** If you are upgrading a Harbor instance with LDAP/AD authentication,
you must make sure **auth_mode** is set to **ldap_auth** in `harbor.cfg` before launching the new version of Harbor. Otherwise, users may not be able to log in after the upgrade.

  - To assist you in migrating the `harbor.cfg` file from v0.5.0 to v1.1.x, a script is provided and described as below. For other versions of Harbor, you need to manually migrate the file `harbor.cfg`.

    ```
    cd harbor
    ./upgrade --source-loc source_harbor_cfg_loc --source-version 0.5.0 --target-loc target_harbor_cfg_loc --target-version 1.1.x
    ```
	**NOTE:** After running the script, make sure you go through `harbor.cfg` to verify all the settings are correct. You can make changes to `harbor.cfg` as needed.

8. Under the directory `./harbor`, run the `./install.sh` script to install the new Harbor instance. If you choose to install Harbor with components like Notary and/or Clair, refer to [Installation & Configuration Guide](../docs/installation_guide.md) for more information.

### Roll back from an upgrade
For any reason, if you want to roll back to the previous version of Harbor, follow the below steps:

1. Stop and remove the current Harbor service if it is still running.

    ```
    cd harbor
    docker-compose down
    ```
2. Restore database from backup file in `/path/to/backup` . 

    ```
    docker run -ti --rm -e DB_USR=root -e DB_PWD=xxxx -v /data/database:/var/lib/mysql -v /path/to/backup:/harbor-migration/backup vmware/harbor-db-migrator:[tag] restore
    ```
    **NOTE:** Rollback from harbor 1.3 to harbor 1.2 should delete `/data/database` directory first, then create new database directory `docker-compose up -d && docker-compose stop`. And must use `vmware/harbor-db-migrator:1.2` to restore. Because of DB engine change.

3. Remove current Harbor instance.
    ```
    rm -rf harbor
    ```

4. Restore the older version package of Harbor.
    ```sh
    mv /my_backup_dir/harbor harbor
    ```

5. Restart Harbor service using the previous configuration.  
   If previous version of Harbor was installed by a release build:
    ```sh
    cd harbor
    ./install.sh
    ```
   **Note:** If you choose to install Harbor with components like Notary and/or Clair, refer to [Installation & Configuration Guide](../docs/installation_guide.md) for more information.

   If your previous version of Harbor was installed from source code:
    ```sh
    cd harbor
    docker-compose up --build -d
    ```

### Migration tool reference
- Use `help` command to show instructions of the migration tool:

    ```docker run --rm -e DB_USR=root -e DB_PWD=xxxx vmware/harbor-db-migrator:[tag] help```

- Use `test` command to test mysql connection:

    ```docker run --rm -e DB_USR=root -e DB_PWD=xxxx -v /data/database:/var/lib/mysql vmware/harbor-db-migrator:[tag] test```
