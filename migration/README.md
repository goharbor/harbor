# Migration guide
Migration is a module for migrating database schema between different version of project [Harbor](https://github.com/vmware/harbor)

This module is for those machine running Harbor's old version, such as 0.1.0. If your Harbor' version is up to date, please ignore this module.

**WARNING!!** You must backup your data before migrating

###Installation
- step 1: 

    ```
    cd migration
    ```
- step 2: change `db_username`, `db_password`, `db_port`, `db_name` in migration.cfg
- step 3: build image from dockerfile

    ```
    docker build -t migrate-tool .
    ```

###Migrate Step
- step 1: stop and remove Harbor service

    ``` 
    docker-compose down
    ```
- step 2: create backup file in `/path/to/backup`

    ```
    docker run -ti --rm -v /data/database:/var/lib/mysql -v /path/to/backup:/harbor-migration/backup migrate-tool backup
    ```

- step 3: perform database schema upgrade

    ```docker run -ti --rm -v /data/database:/var/lib/mysql migrate-tool up head```



- step 4: rebuild newest Harbor images and restart service

    ```
    docker-compose build && docker-compose up -d
    ```

You may change `/data/database` to the mysql volumes path you set in docker-compose.yml.

###Migration operation reference
- You can use `help` to show instruction of Harbor migration

    ```docker run migrate-tool help```
    
- You can use `test` to test mysql connection in Harbor migration

    ```docker run --rm -v /data/database:/var/lib/mysql migrate-tool test```

- You can restore from backup file in `/path/to/backup`

    ```
    docker run -ti --rm -v /data/database:/var/lib/mysql -v /path/to/backup:/harbor-migration/backup migrate-tool restore
    ```
