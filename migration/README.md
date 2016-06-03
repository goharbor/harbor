# Migration guide
Migration is a module for migrating database schema between different version of project [Harbor](https://github.com/vmware/harbor)

This module is for those machine running Harbor's old version, such as 0.1.0. If your harbor' version is up to date, please ignore this module.

**WARNING!!** You must backup your data before migrating

###Installation
- step 1: change `db_username`, `db_password`, `db_port`, `db_name` in migration.cfg
- step 2: build image from dockerfile
    ```
    cd harbor-migration
    
    docker build -t your-image-name .
    ```

you may change `/data/database` to the mysql volumes path you set in docker-compose.yml.
###Migrate Step
- step 1: stop and remove harbor service

    ``` 
    docker-compose down
    ```
- step 2: create backup file in `/path/to/backup`

    ```
    docker run -ti -v /data/database:/var/lib/mysql -v /path/to/backup:/harbor-migration/backup migrate-tool backup
    ```

- step 3: perform database schema upgrade

    ```docker run -ti -v /data/database:/var/lib/mysql migrate-tool up head```



- step 4: rebuild newest harbor images and restart service

    ```
    docker-compose build && docker-compose up -d
    ```

###All migration operation
- show instruction of harbor-migration

    ```docker run migrate-tool help```
    
- test mysql connection in harbor-migration

    ```docker run -v /data/database:/var/lib/mysql migrate-tool test```

- create backup file in `/path/to/backup`

    ```
    docker run -ti -v /data/database:/var/lib/mysql -v /path/to/backup:/harbor-migration/backup migrate-tool backup
    ```

- restore from backup file in `/path/to/backup`

    ```
    docker run -ti -v /data/database:/var/lib/mysql -v /path/to/backup:/harbor-migration/backup migrate-tool restore
    ```

- perform database schema upgrade

    ```docker run -ti -v /data/database:/var/lib/mysql migrate-tool up head```
