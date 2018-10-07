# Harbor upgrade and database migration guide

When upgrading your existing Habor instance to a newer version, you may need to migrate the data in your database and the settings in harbor.cfg. Refer to [change log](../tools/migration/changelog.md) to find out whether there is any change in the database. If there is, you should go through the database migration process. Since the migration may alter the database schema and the settings of harbor.cfg, you should **always** back up your data before any migration.

*If your install Harbor for the first time, or the database version is the same as that of the lastest version, you do not need any database migration.*

**NOTE:**
- From v1.6.0 on, Harbor migrates DB from MariaDB to Postgresql, and combines Harbor, Notary and Clair DB into one. 

- From v1.5.0 on, the migration tool add support for the harbor.cfg migration, which supports upgrade from v1.2.x, v1.3.x and v1.4.x.

- From v1.2 on, you need to use the release version as the tag of the migrator image. 'latest' is no longer used for new release.

- You must back up your data before any data migration.

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

3. Get the latest Harbor release package from Github:
   https://github.com/goharbor/harbor/releases

4. Before upgrading Harbor, perform migration first.  The migration tool is delivered as a docker image, so you should pull the image from docker hub. Replace [tag] with the release version of Harbor (e.g. v1.5.0) in the below command:

    **NOTE:** Before harbor 1.5 , image name of the migration tool is `goharbor/harbor-db-migrator:[tag]`

    ```
    docker pull goharbor/harbor-migrator:[tag]
    ```

5. Back up database/harbor.cfg to a directory such as `/path/to/backup`. You need to create the directory if it does not exist.  Also, note that the username and password to access the db are provided via environment variable "DB_USR" and "DB_PWD". 

    **NOTE:** Upgrade from harbor 1.2 or older to harbor 1.3 must use `goharbor/harbor-db-migrator:1.2`. Because DB engine replaced by MariaDB in harbor 1.3

    ```
    docker run -it --rm -e DB_USR=root -e DB_PWD={db_pwd} -v ${harbor_db_path}:/var/lib/mysql -v ${harbor_cfg}:/harbor-migration/harbor-cfg/harbor.cfg -v ${backup_path}:/harbor-migration/backup goharbor/harbor-migrator:[tag] backup
    ```

    **NOTE:** By default, the migrator handles the backup for DB and CFG. If you want to backup DB or CFG only, refer to the following commands.
    
    ```
    docker run -it --rm -e DB_USR=root -e DB_PWD={db_pwd} -v ${harbor_db_path}:/var/lib/mysql -v ${backup_path}:/harbor-migration/backup goharbor/harbor-migrator:[tag] --db backup
    ```

    ```
    docker run -it --rm -v ${harbor_cfg}:/harbor-migration/harbor-cfg/harbor.cfg -v ${backup_path}:/harbor-migration/backup goharbor/harbor-migrator:[tag] --cfg backup
    ```    

6.  Upgrade database schema, harbor.cfg and migrate data.

    **NOTE:** In v1.6.0, you needs to DO three sequential steps to fully migrate Harbor, Notary and Clair's DB. The migration of Notary and Clair's DB depends on Harbor's DB, you need to first upgrade Harbor's DB, then upgrade Notary and Clair's DB. The following command handles the upgrade for Harbor DB and CFG, not include Notary and Clair DB. 

    ```
    docker run -it --rm -e DB_USR=root -e DB_PWD={db_pwd} -v ${harbor_db_path}:/var/lib/mysql -v ${harbor_cfg}:/harbor-migration/harbor-cfg/harbor.cfg goharbor/harbor-migrator:[tag] up
    ```

    **NOTE:** You must run migration of Notary and Clair's DB before launch Harbor. If you want to upgrade Notary and Clair DB, refer to the following commands:

    ```
    docker run -it --rm -e DB_USR=root -v /data/notary-db/:/var/lib/mysql -v /data/database:/var/lib/postgresql/data goharbor/harbor-migrator:${tag} --db up

    docker run -it --rm -v /data/clair-db/:/clair-db -v /data/database:/var/lib/postgresql/data goharbor/harbor-migrator:${tag} --db up
    ```

    **NOTE:** If you want to upgrade DB or CFG only, refer to the following commands:
    
    ```
    docker run -it --rm -e DB_USR=root -e DB_PWD={db_pwd} -v ${harbor_db_path}:/var/lib/mysql goharbor/harbor-migrator:[tag] --db up

    docker run -it --rm -e DB_USR=root -v /data/notary-db/:/var/lib/mysql -v /data/database:/var/lib/postgresql/data goharbor/harbor-migrator:${tag} --db up

    docker run -it --rm -v /data/clair-db/:/clair-db -v /data/database:/var/lib/postgresql/data goharbor/harbor-migrator:${tag} --db up
    ```

    **NOTE:** The ${harbor_cfg} will be overwritten, you must move it to your installation directory after migration.

    ```
    docker run -it --rm -v ${harbor_cfg}:/harbor-migration/harbor-cfg/harbor.cfg goharbor/harbor-migrator:[tag] --cfg up
    ```

     **NOTE:** Some errors like

    ```
    [ERROR] Missing system table mysql.roles_mapping; please run mysql_upgrade to create it
    [ERROR] Incorrect definition of table mysql.event: expected column 'sql_mode' at position ... ...
    [ERROR] mysqld: Event Scheduler: An error occurred when initializing system tables. Disabling the Event Scheduler.
    [Warning] Failed to load slave replication state from table mysql.gtid_slave_pos: 1146: Table 'mysql.gtid_slave_pos' doesn't exist
    ```
    will be occurred during upgrading from harbor 1.2 to harbor 1.3, just ignore them if harbor can start successfully.

    ```
    /usr/lib/python2.7/site-packages/psycopg2/__init__.py:144: UserWarning: The psycopg2 wheel package will be renamed from release 2.8; in order to keep installing from binary please use "pip install psycopg2-binary" instead. For details see: <http://initd.org/psycopg/docs/install.html#binary-install-from-pypi>.
    ```
    will be occurred during upgrading from harbor <= v1.5.0 to harbor v1.6.0, just ignore them if harbor can start successfully.

7. Under the directory `./harbor`, run the `./install.sh` script to install the new Harbor instance. If you choose to install Harbor with components like Notary and/or Clair, refer to [Installation & Configuration Guide](../docs/installation_guide.md) for more information.

### Roll back from an upgrade
For any reason, if you want to roll back to the previous version of Harbor, follow the below steps:

**NOTE:** Roll back doesn't support upgrade across v1.5.0, like from v1.2.0 to v1.6.0. It's because Harbor changes DB to Postgresql from v1.6.0, the migrator cannot roll back data to MariaDB.    

1. Stop and remove the current Harbor service if it is still running.

    ```
    cd harbor
    docker-compose down
    ```
2. Restore database from backup file in `/path/to/backup` . 

    ```
    docker run -it --rm -e DB_USR=root -e DB_PWD={db_pwd} -v ${harbor_db_path}:/var/lib/mysql -v ${harbor_cfg}:/harbor-migration/harbor-cfg/harbor.cfg -v ${backup_path}:/harbor-migration/backup goharbor/harbor-migrator:[tag] restore
    ```
 
    **NOTE:** By default, the migrator handles the restore for DB and CFG. If you want to restore DB or CFG only, 
    refer to the following commands:
    
    ```
    docker run -it --rm -e DB_USR=root -e DB_PWD={db_pwd} -v ${harbor_db_path}:/var/lib/mysql -v ${backup_path}:/harbor-migration/backup goharbor/harbor-migrator:[tag] --db restore
    ```

    ```
    docker run -it --rm -v ${harbor_cfg}:/harbor-migration/harbor-cfg/harbor.cfg -v ${backup_path}:/harbor-migration/backup goharbor/harbor-migrator:[tag] --cfg restore
    ```
 
    **NOTE:** Rollback from harbor 1.3 to harbor 1.2 should delete `/data/database` directory first, then create new database directory `docker-compose up -d && docker-compose stop`. And must use `goharbor/harbor-db-migrator:1.2` to restore. Because of DB engine change.

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
- Use `test` command to test mysql connection:

    ```docker run -it --rm -e DB_USR=root -e DB_PWD={db_pwd} -v ${harbor_db_path}:/var/lib/mysql -v ${harbor_cfg}:/harbor-migration/harbor-cfg/harbor.cfg goharbor/harbor-migrator:[tag] test```
