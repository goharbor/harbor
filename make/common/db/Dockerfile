FROM vmware/mariadb-photon:10.2.8

COPY registry.sql /docker-entrypoint-initdb.d/
COPY registry-flag.sh /docker-entrypoint-initdb.d/
COPY upgrade.sh /docker-entrypoint-updatedb.d/

