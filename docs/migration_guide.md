# Harbor upgrade and database migration guide

When upgrading your existing Habor instance to a newer version, you may need to migrate the data in your database. Refer to [change log](../migration/changelog.md) to find out whether there is any change in the database. If there is, you should go through the database migration process. Since the migration may alter the database schema, you should **always** back up your data before any migration.

*If your install Harbor for the first time, or the database version is the same as that of the lastest version, you do not need any database migration.*

**NOTE:** From 1.2, you need to use release version as the tag of migrator image. 'latest' is no longer used for new release.

**NOTE:** You must backup your data before any data migration.

### Upgrading Harbor and migrating data

1. Log in to the host that Harbor runs on, stop and remove existing Harbor instance if it is still running:

    ```
    cd harbor
    docker-compose down
    ```

2.  Back up Harbor's current files so that you can roll back to the current version when it is necessary.
    ```sh
    cd ..
    mv harbor /tmp/harbor
    ```

3. Get the lastest Harbor release package from Github:
   https://github.com/vmware/harbor/releases

4. Before upgrading Harbor, perform database migration first.  The migration tool is delivered as a docker image, so you should pull the image from docker hub:

    ```
    docker pull vmware/harbor-db-migrator:[tag]
    ```

5. Back up database to a directory such as `/path/to/backup`. You need to create the directory if it does not exist.  Also, note that the username and password to access the db are provided via environment variable "DB_USR" and "DB_PWD"

    ```
    docker run -ti --rm -e DB_USR=root -e DB_PWD=xxxx -v /data/database:/var/lib/mysql -v /path/to/backup:/harbor-migration/backup vmware/harbor-db-migrator:[tag] backup
    ```

6.  Upgrade database schema and migrate data:

    ```
    docker run -ti --rm -e DB_USR=root -e DB_PWD=xxxx -v /data/database:/var/lib/mysql vmware/harbor-db-migrator:[tag] up head
    ```

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

8. Under the directory `./harbor`, run the `./install.sh` script to install the new Harbor instance.

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

3. Remove current Harbor instance.
    ```
    rm -rf harbor
    ```

4. Restore the older version package of Harbor.
    ```sh
    mv /tmp/harbor harbor
    ```

5. Restart Harbor service using the previous configuration.  
   If previous version of Harbor was installed by a release build:
    ```sh
    cd harbor
    ./install.sh
    ```

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
