# Harbor upgrade and migration guide

This guide only covers upgrade and migration to version >= v1.8.0

When upgrading your existing Harbor instance to a newer version, you may need to migrate the data in your database and the settings in `harbor.cfg`. 
Since the migration may alter the database schema and the settings of `harbor.cfg`, you should **always** back up your data before any migration.

**NOTE:**

- Again, you must back up your data before any data migration.

- This guide only covers the migration from v1.6.0 to current version, if you are upgrading from earlier versions please 
refer to the migration guide in release branch to upgrade to v1.6.0 and follow this guide to do the migration to later version. 

- From v1.6.0 on, Harbor will automatically try to do the migrate the DB schema when it starts, so if you are upgrading from v1.6.0 
or above it's not necessary to call the migrator tool to migrate the schema.

- For the change in Database schema please refer to [change log](../tools/migration/db/changelog.md).

- Since v1.8.0, the configuration of Harbor has changed to `.yml` file, the migrator will transform the configuration 
file from `harbor.cfg` to `harbor.yml`.  The command will be a little different to perform this migration, please make sure
you follow the steps below.


### Upgrading Harbor and migrating data

1. Log in to the host that Harbor runs on, stop and remove existing Harbor instance if it is still running:
    ```
    cd harbor
    docker-compose down
    ```

2.  Back up Harbor's current files so that you can roll back to the current version when it is necessary.
    ```
    mv harbor /my_backup_dir/harbor
    ```
    Back up database (by default in directory `/data/database`)
    ```
    cp -r /data/database /my_backup_dir/
    ```

3. Get the latest Harbor release package from Github:
   https://github.com/goharbor/harbor/releases

4. Before upgrading Harbor, perform migration first.  The migration tool is delivered as a docker image, so you should pull the image from docker hub. Replace [tag] with the release version of Harbor (e.g. v1.5.0) in the below command:
    ```
    docker pull goharbor/harbor-migrator:[tag]
    ```

5. Upgrade from `harbor.cfg` to `harbor.yml`
    **NOTE:** You can find the ${harbor_yml} in the extracted installer you got in step `3`, after the migration the file `harbor.yml` 
    in that path will be updated with the values from ${harbor_cfg}
    
    ```
    docker run -it --rm -v ${harbor_cfg}:/harbor-migration/harbor-cfg/harbor.cfg -v ${harbor_yml}:/harbor-migration/harbor-cfg-out/harbor.yml goharbor/harbor-migrator:[tag] --cfg up
    ```
    **NOTE:** The schema upgrade and data migration of Database is performed by core when Harbor starts, if the migration fails,
    please check the log of core to debug.

6. Under the directory `./harbor`, run the `./install.sh` script to install the new Harbor instance. If you choose to install Harbor with components like Notary, Clair, and chartmuseum, refer to [Installation & Configuration Guide](../docs/installation_guide.md) for more information.


### Roll back from an upgrade
For any reason, if you want to roll back to the previous version of Harbor, follow the below steps:

**NOTE:** Roll back doesn't support upgrade across v1.5.0, like from v1.2.0 to v1.7.0. This is because Harbor changes DB to PostgreSQL from v1.7.0, the migrator cannot roll back data to MariaDB.    

1. Stop and remove the current Harbor service if it is still running.
    ```
    cd harbor
    docker-compose down
    ```
    
2. Remove current Harbor instance.
    ```
    rm -rf harbor
    ```
    
3. Restore the older version package of Harbor.
    ```sh
    mv /my_backup_dir/harbor harbor
    ```
    
4. Restore database, copy the data files from backup directory to you data volume, by default `/data/database`. 

5. Restart Harbor service using the previous configuration.  
   If previous version of Harbor was installed by a release build:
    ```sh
    cd harbor
    ./install.sh
    ```
