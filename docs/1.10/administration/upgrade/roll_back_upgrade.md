[Back to table of contents](../../index.md)

----------

# Roll Back from an Upgrade

If, for any reason, you need to roll back to the previous version of Harbor, perform the following steps.

1. Stop and remove the current Harbor service if it is still running.

    ```sh
    cd harbor
    docker-compose down
    ```

2. Remove current Harbor instance.

    ```sh
    rm -rf harbor
    ```

3. Restore the older version of Harbor.

    ```sh
    mv /my_backup_dir/harbor harbor
    ```

4. To restore the database, copy the data files from the backup directory to your data volume, which by default is `/data/database`.

5. Restart the Harbor service using the previous configuration.  
   
   If the previous version of Harbor was installed by a release build:

    ```sh
    cd harbor
    ./install.sh
    ```

**NOTE**: While you can roll back an upgrade to the state before you started the upgrade, Harbor does not support downgrades.

----------

[Back to table of contents](../../index.md)