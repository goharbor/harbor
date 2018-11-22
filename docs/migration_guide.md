# Harbor upgrade and database migration guide

Refer to [change log](../tools/migration/changelog.md) to find out whether there is any change in the database or harbor.cfg. If there is, you should go through the database migration process.

**NOTE:**
- From v1.6.0 onwards, Harbor migrated from MariaDB to Postgresql, combining Harbor, Notary and Clair databases into one. 

- From v1.5.0 onwards, the migration tool adds support for harbor.cfg migration, which supports upgrading from v1.2.x, v1.3.x and v1.4.x.

- From v1.2 onwards, you need to use the release version as the tag of the migrator image. 'latest' is no longer used for new releases.

- Migrations may alter the database schema and harbor.cfg therefore, **always** back up your data before any migration.

- If you install Harbor for the first time, or the database version is the same as that of the lastest version, you do not need any database migration.

## Upgrading Harbor

**NOTE:** We assume Harbor is installed under `/opt/harbor`, subsequent commands will be based off this path

1. Login to the host that Harbor runs on, and stop all Harbor related containers

    ```sh
    cd /opt/harbor
    docker-compose down
    ```

2. Pull the migration tool docker image, replace [tag] to the version of Harbor that you are upgrading to

    **NOTE:** Before Harbor 1.5 , the image name is `goharbor/harbor-db-migrator:[tag]`

    ```
    docker pull goharbor/harbor-migrator:[tag]
    ```

3. Back up the whole of Harbor contents

    ```sh
    mkdir /opt/harbor_full_bkup
    cp -r /opt/harbor/* /opt/harbor_full_bkup/
    ```

4. We can also back up either Harbor's database or harbor.cfg using the migration tool

    The following environment variables will be used in subsequent commands. The values shown are for demonstration

    ```sh
    export db_pwd=examplepassword
    export harbor_backup=/opt/harbor_bkup
    export harbor_db=/opt/harbor/data/database
    export harbor_notary_db=/opt/harbor/data/notary-db
    export harbor_clair_db=/opt/harbor/data/clair-db
    export harbor_cfg=/opt/harbor/harbor.cfg
    ```

    Back up both Harbor database and harbor.cfg

    **NOTE:** From Harbor version 1.2 to 1.3 use `goharbor/harbor-db-migrator:1.2`. As the DB engine is replaced with MariaDB in harbor 1.3

    ```
    docker run -it --rm -e DB_USR=root -e DB_PWD=${db_pwd} -v ${harbor_db}:/var/lib/mysql -v ${harbor_cfg}:/harbor-migration/harbor-cfg/harbor.cfg -v ${harbor_backup}:/harbor-migration/backup goharbor/harbor-migrator:[tag] backup
    ```

    Back up only Harbor database
    ```
    docker run -it --rm -e DB_USR=root -e DB_PWD=${db_pwd} -v ${harbor_db}:/var/lib/mysql -v ${harbor_backup}:/harbor-migration/backup goharbor/harbor-migrator:[tag] --db backup
    ```

    Back up only Harbor harbor.cfg
    ```
    docker run -it --rm -v ${harbor_cfg}:/harbor-migration/harbor-cfg/harbor.cfg -v ${harbor_backup}:/harbor-migration/backup goharbor/harbor-migrator:[tag] --cfg backup
    ```    

5.  Upgrade database schema, harbor.cfg and migrate data.

    **NOTE:** For Harbor v1.6.0, the migration consists of three steps that must be performed sequentially. As the migration of Notary and Clair's DB depends on Harbor's DB, therefore we need to first upgrade Harbor's DB, followed by Notary and finally Clair's DB. 
    
    Upgrade both Harbor DB and harbor.cfg, not including Notary and Clair DB.
    ```
    docker run -it --rm -e DB_USR=root -e DB_PWD=${db_pwd} -v ${harbor_db}:/var/lib/mysql -v ${harbor_cfg}:/harbor-migration/harbor-cfg/harbor.cfg goharbor/harbor-migrator:[tag] up
    ```

    **NOTE:** Ensure migration of Notary and Clair's DB is done before launching Harbor. For upgrading Notary and Clair DB only, refer to the following commands:

    ```
    docker run -it --rm -e DB_USR=root -v ${harbor_notary_db}:/var/lib/mysql -v ${harbor_db}:/var/lib/postgresql/data goharbor/harbor-migrator:[tag] --db upharbor_db

    docker run -it --rm -v ${harbor_clair_db}:/clair-db -v ${harbor_db}:/var/lib/postgresql/data goharbor/harbor-migrator:[tag] --db up
    ```

    **NOTE:** For upgrading DB or CFG only, refer to the following commands:
    
    ```
    docker run -it --rm -e DB_USR=root -e DB_PWD=${db_pwd} -v ${harbor_db}:/var/lib/mysql goharbor/harbor-migrator:[tag] --db up

    docker run -it --rm -e DB_USR=root -v ${harbor_notary_db}:/var/lib/mysql -v ${harbor_db}:/var/lib/postgresql/data goharbor/harbor-migrator:[tag] --db up

    docker run -it --rm -v ${harbor_clair_db}:/clair-db -v  ${harbor_db}:/var/lib/postgresql/data goharbor/harbor-migrator:[tag] --db up
    ```

    **NOTE:** ${harbor_cfg} will be overwritten, you must move it to your installation directory after migration.

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
    may occurr during upgrading from harbor 1.2 to harbor 1.3, ignore them if harbor can still start successfully.

    ```
    /usr/lib/python2.7/site-packages/psycopg2/__init__.py:144: UserWarning: The psycopg2 wheel package will be renamed from release 2.8; in order to keep installing from binary please use "pip install psycopg2-binary" instead. For details see: <http://initd.org/psycopg/docs/install.html#binary-install-from-pypi>.
    ```
    may occurr during upgrading from harbor <= v1.5.0 to harbor v1.6.0, ignore them if harbor can still start successfully.

6. Download either the online/offline release of the version of Harbor that you are upgrading to, and extract the contents into the original Harbor directory

7. Under the directory `./harbor`, run `./install.sh` script to install the new Harbor instance. If you choose to install Harbor with other components , refer to [Installation & Configuration Guide](../docs/installation_guide.md) for more information.

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
