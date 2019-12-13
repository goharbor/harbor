# Reconfigure Harbor and Manage the Harbor Lifecycle 

You can use `docker-compose` to manage the lifecycle of Harbor. Some useful commands are listed below. You must run the commands in the same directory as `docker-compose.yml`.

### Stop Harbor:

``` sh
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

### Restart Harbor after Stopping:

``` sh
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

### Reconfigure Harbor

To reconfigure Harbor, stop the existing Harbor instance and update `harbor.yml`. Then run `prepare` script to populate the configuration. Finally re-create and start the Harbor instance.

``` sh
$ sudo docker-compose down -v
$ vim harbor.yml
$ sudo prepare
$ sudo docker-compose up -d
```

### Other Commands

Remove Harbor's containers while keeping the image data and Harbor's database files on the file system:

``` sh
$ sudo docker-compose down -v
```

Remove Harbor's database and image data for a clean re-installation:

``` sh
$ rm -r /data/database
$ rm -r /data/registry
```

### Managing the Harbor Lifecycle  with Notary, Clair and Chart Repository Service

If you want to install Notary, Clair and chart repository service together, you should include all the components in the prepare commands:

``` sh
$ sudo docker-compose down -v
$ vim harbor.yml
$ sudo prepare --with-notary --with-clair --with-chartmuseum
$ sudo docker-compose up -d
```

Please check the [Docker Compose command-line reference](https://docs.docker.com/compose/reference/) for more on docker-compose.

