# Harbor Upgrade and Migration Guide

This guide covers upgrade and migration to version 1.9.0. This guide only covers migration from v1.7.x and later to the current version. If you are upgrading from an earlier version, refer to the migration guide in the `release-1.7.0` branch to upgrade to v1.7.x first, then follow this guide to perform the migration to this version.

When upgrading an existing Harbor 1.7.x instance to a newer version, you might need to migrate the data in your database and the settings in `harbor.cfg`.
Since the migration might alter the database schema and the settings of `harbor.cfg`, you should **always** back up your data before any migration.

**NOTES:**

- Again, you must back up your data before any data migration.
- Since v1.8.0, the configuration of Harbor has changed to a `.yml` file. If you are upgrading from 1.7.x, the migrator will transform the configuration file from `harbor.cfg` to `harbor.yml`. The command will be a little different to perform this migration, so make sure you follow the steps below.
- In version 1.9.0, some containers are started by `non-root`. This does not pose problems if you are upgrading an officially released version of Harbor, but if you have deployed a customized instance of Harbor, you might encounter permission issues.
- In previous releases, user roles took precedence over group roles in a project. In this version, user roles and group roles are combined so that the user has whichever set of permissions is highest. This might cause the roles of certain users to change during upgrade.
- With the introduction of storage and artifact quotas in version 1.9.0, migration from 1.7.x and 1.8.x might take a few minutes. This is because the `core` walks through all blobs in the registry and populates the database with information about the layers and artifacts in projects.
- With the introduction of storage and artifact quotas in version 1.9.0, replication between version 1.9.0 and a previous version of Harbor does not work. You must upgrade all Harbor nodes to 1.9.0 if you have configured replication between them.

## Upgrading Harbor and Migrating Data

#### Shut Down the Harbor Service

Log in to the host on which Harbor is running. Stop the existing Harbor instance if it is still running, and delete it:

```sh
cd harbor
docker-compose down
```

#### Backup Harbor Data

Back up Harbor's data, so that you can roll back to the current version if necessary.

```sh
mv harbor /my_backup_dir/harbor
```

Back up database (by default in directory `/data/database`)

```sh
cp -r /data/database /my_backup_dir/
```

#### Get Package of Latest Harbor Installer

Download the latest version of Harbor from [https://github.com/goharbor/harbor/releases](https://github.com/goharbor/harbor/releases).

#### Migrating Configuration

Before you upgrade Harbor, you must first perform migration. The migration tool is delivered as a Docker image.

1. **Online installer** Pull the migration image from Docker Hub. Replace `[tag]` with the Harbor release version, for example, `v1.5.0`, in the following command:

    ```sh
    docker pull goharbor/harbor-migrator:[tag]
    ```

    **Offline installer** The offline installer loads the migration image from the tarball that is included in the offline installer package. Replace `[version]` with the Harbor release version, for example, `v1.5.0`, in the following command:

    ```sh
    tar zxf <offline package>
    docker image load -i harbor/harbor.[version].tar.gz
    ```

2. **If your current Harbor version is v1.7.x or earlier**, you must migrate the configuration file from `harbor.cfg` to `harbor.yml`. You must mount both `harbor.cfg` and `harbor.yml` to be able to run the command.

    **NOTE:** After the migration the `harbor.yml` is updated with the values from `${harbor_cfg}`

    ```sh
    docker run -it --rm -v ${harbor_cfg}:/harbor-migration/harbor-cfg/harbor.yml -v ${harbor_yml}:/harbor-migration/harbor-cfg-out/harbor.yml goharbor/harbor-migrator:[tag] --cfg up
    ```

    **If your current Harbor version is v1.8.0 or above**, there is no `harbor.cfg` file. So, you only need to mount `harbor.yml`.

    **NOTE:** This command only updates ` ${harbor_yml}`. Manually copy this file to the correct location after the upgrade.

    ```sh
    docker run -it --rm -v ${harbor_yml}:/harbor-migration/harbor-cfg/harbor.yml goharbor/harbor-migrator:[tag] --cfg up
    ```

    **NOTE:** The schema upgrade and data migration of the database is performed by core when Harbor starts. If the migration fails, check the core log to debug.

#### Install Harbor

Run `./install.sh` in the `./harbor` directory to install the new Harbor instance. For instructions about how to install Harbor with components such as Notary, Clair, and chartmuseum, see the [Installation & Configuration Guide](../docs/installation_guide.md)

## Roll Back from an Upgrade

If, for any reason, you want to roll back to the previous version of Harbor, perform the following steps:

1. Stop and remove the current Harbor service if it is still running.

    ```sh
    cd harbor
    docker-compose down
    ```

2. Remove current Harbor instance.

    ```sh
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

**NOTE**: While you can roll back an upgrade to the state before you started the upgrade, Harbor does not support downgrades.
