# migration
Migration is a module for migrating database schema between different version of project [harbor](https://github.com/vmware/harbor)

**WARNING!!** You must backup your data before migrating

###installation
- step 1: modify migration.cfg
- step 2: build image from dockerfile
    ```
    cd harbor-migration
    
    docker build -t your-image-name .
    ```

###migration operation
- show instruction of harbor-migration

    ```docker run your-image-name help```

- create backup file in `/path/to/backup`

    ```
    docker run -ti -v /data/database:/var/lib/mysql -v /path/to/backup:/harbor-migration/backup your-image-name backup
    ```

- restore from backup file in `/path/to/backup`

    ```
    docker run -ti -v /data/database:/var/lib/mysql -v /path/to/backup:/harbor-migration/backup your-image-name restore
    ```

- perform database schema upgrade

    ```docker run -ti -v /data/database:/var/lib/mysql your-image-name up head```

- perform database schema downgrade(downgrade has been disabled)

    ```docker run -v /data/database:/var/lib/mysql your-image-name down base```

###migration step
- step 1: stop and remove harbor service

    ``` 
    docker-compose stop && docker-compose rm -f
    ```
- step 2: perform migration operation
- step 3: rebuild newest harbor images and restart service

    ```
    docker-compose build && docker-compose up -d
    ```
