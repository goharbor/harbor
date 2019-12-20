[Back to table of contents](../_index.md)

----------

# Reconfigure Harbor and Manage the Harbor Lifecycle 

You use `docker-compose` to manage the lifecycle of Harbor. This topic provides some useful commands. You must run the commands in the directory in which `docker-compose.yml` is located.

See the [Docker Compose command-line reference](https://docs.docker.com/compose/reference/) for more information about `docker-compose`.

## Stop Harbor

To stop Harbor, run the following command.

```
$ sudo docker-compose stop
Stopping nginx              ... done
Stopping harbor-portal      ... done
Stopping harbor-jobservice  ... done
Stopping harbor-core        ... done
Stopping registry           ... done
Stopping redis              ... done
Stopping registryctl        ... done
Stopping harbor-db          ... done
Stopping harbor-log         ... done
```

## Restart Harbor 

To restart Harbor, run the following command.

```
$ sudo docker-compose start
Starting log         ... done
Starting registry    ... done
Starting registryctl ... done
Starting postgresql  ... done
Starting core        ... done
Starting portal      ... done
Starting redis       ... done
Starting jobservice  ... done
Starting proxy       ... done
```

## Reconfigure Harbor

To reconfigure Harbor, perform the following steps.

1. Stop Harbor. 

    ```
    sudo docker-compose down -v
    ```
1. Update `harbor.yml`. 

    ```
    vim harbor.yml
    ```
1. Run the `prepare` script to populate the configuration.

    ```
    sudo prepare
    ```
    To reconfigure Harbor to install Notary, Clair, and the chart repository service, include all of the components in the `prepare` command.

    ```
    sudo prepare --with-notary --with-clair --with-chartmuseum
    ```
1. Re-create and start the Harbor instance.

    ```
    sudo docker-compose up -d
    ```

## Other Commands

Remove Harbor's containers but keep all of the image data and Harbor's database files in the file system:

```
$ sudo docker-compose down -v
```

Remove the Harbor database and image data before performing a clean re-installation:

```
$ rm -r /data/database
$ rm -r /data/registry
```

----------

[Back to table of contents](../_index.md)