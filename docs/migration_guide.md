# Harbor Upgrade and Migration Guide

This guide covers upgrade and migration to versions >= v1.8.0. 

When upgrading an existing Harbor 1.7.x instance to a newer version, you might need to migrate the data in your database and the settings in `harbor.cfg`. 
Since the migration might alter the database schema and the settings of `harbor.cfg`, you should **always** back up your data before any migration.

**NOTES:**

- Again, you must back up your data before any data migration.

- This guide only covers migration from v1.7.x to the current version. If you are upgrading from earlier versions please refer to the migration guide in the relevant release branch to upgrade to v1.7.x first, then follow this guide to perform the migration to this version. 

- Since v1.8.0, the configuration of Harbor has changed to a `.yml` file. If you are upgrading from 1.7.x, the migrator will transform the configuration file from `harbor.cfg` to `harbor.yml`. The command will be a little different to perform this migration, so make sure you follow the steps below.

### Upgrading Harbor and Migrating Data

1. Log in to the host that Harbor runs on, stop and remove existing Harbor instance if it is still running:
    ```
    cd harbor
    docker-compose down
    ```

2.  Back up Harbor's current files so that you can roll back to the current version if necessary.
    ```
    mv harbor /my_backup_dir/harbor
    ```
    Back up database (by default in directory `/data/database`)
    ```
    cp -r /data/database /my_backup_dir/
    ```

3. Get the latest Harbor release package from Github:
   https://github.com/goharbor/harbor/releases

4. Before upgrading Harbor, perform migration first.  The migration tool is delivered as a docker image, so you should pull the image from docker hub. Replace [tag] with the release version of Harbor (for example, v1.9.0) in the command below:
    ```
    docker pull goharbor/harbor-migrator:[tag]
    ```

5. If you are upgrading from v1.7.x, migrate from `harbor.cfg` to `harbor.yml`.
    **NOTE:** You can find the ${harbor_yml} in the extracted installer you got in step `3`, after the migration the file `harbor.yml` 
    in that path will be updated with the values from ${harbor_cfg}
    
    ```
    docker run -it --rm -v ${harbor_cfg}:/harbor-migration/harbor-cfg/harbor.yml -v ${harbor_yml}:/harbor-migration/harbor-cfg-out/harbor.yml goharbor/harbor-migrator:[tag] --cfg up
    ```
    **NOTE:** The schema upgrade and data migration of the database is performed by core when Harbor starts, if the migration fails, please check the log of core to debug.

6. Under the directory `./harbor`, run the `./install.sh` script to install the new Harbor instance. If you choose to install Harbor with components such as Notary, Clair, and chartmuseum, refer to [Installation & Configuration Guide](../docs/installation_guide.md) for more information.


### Roll back from an upgrade
If, for any reason, you want to roll back to the previous version of Harbor, follow the below steps: 

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
