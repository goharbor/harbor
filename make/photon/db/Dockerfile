FROM vmware/mariadb-photon:__version__

HEALTHCHECK CMD mysqladmin -uroot -p$MYSQL_ROOT_PASSWORD ping

COPY registry.sql /docker-entrypoint-initdb.d/
COPY registry-flag.sh /docker-entrypoint-initdb.d/
COPY upgrade.sh /docker-entrypoint-updatedb.d/

