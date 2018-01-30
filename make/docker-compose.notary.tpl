version: '2'
services:
  ui:
    networks:
      - harbor-notary
  proxy:
    networks:
      - harbor-notary
  notary-server:
    image: vmware/notary-server-photon:__notary_version__
    container_name: notary-server
    restart: always
    networks:
      - notary-mdb
      - notary-sig
      - harbor-notary
    volumes:
      - ./common/config/notary:/config:z
    depends_on:
      - notary-db
      - notary-signer
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "notary-server"
  notary-signer:
    image: vmware/notary-signer-photon:__notary_version__
    container_name: notary-signer
    restart: always
    networks:
      notary-mdb:
      notary-sig:
        aliases:
          - notarysigner
    volumes:
      - ./common/config/notary:/config:z
    env_file:
      - ./common/config/notary/signer_env
    depends_on:
      - notary-db
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "notary-signer"
  notary-db:
    image: vmware/mariadb-photon:__mariadb_version__
    container_name: notary-db
    restart: always
    networks:
      notary-mdb:
        aliases:
          - mysql
    volumes:
      - ./common/config/notary/mysql-initdb.d:/docker-entrypoint-initdb.d:z
      - /data/notary-db:/var/lib/mysql:z
    environment:
      - TERM=dumb
      - MYSQL_ALLOW_EMPTY_PASSWORD="true"
    command: mysqld --innodb_file_per_table
    depends_on:
      - log
    logging:
      driver: "syslog"
      options:  
        syslog-address: "tcp://127.0.0.1:1514"
        tag: "notary-db"
networks:
  harbor-notary:
    external: false
  notary-mdb:
    external: false
  notary-sig:
    external: false
