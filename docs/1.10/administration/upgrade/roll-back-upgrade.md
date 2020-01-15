---
title: Roll Back from an Upgrade
---

If, for any reason, you want to roll back to the previous version of Harbor, perform the following steps:

1. Stop and remove the current Harbor service if it is still running.

    ```shell
    cd harbor
    docker-compose down
    ```

2. Remove current Harbor instance.

    ```shell
    rm -rf harbor
    ```

3. Restore the older version package of Harbor.

    ```shell
    mv /my_backup_dir/harbor harbor
    ```

4. Restore database, copy the data files from backup directory to you data volume, by default `/data/database`.

5. Restart Harbor service using the previous configuration.  
   If previous version of Harbor was installed by a release build:

    ```shell
    cd harbor
    ./install.sh
    ```

**NOTE**: While you can roll back an upgrade to the state before you started the upgrade, Harbor does not support downgrades.