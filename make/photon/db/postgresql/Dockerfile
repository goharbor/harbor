FROM vmware/postgresql-photon:__postgresql_version__

COPY registry.sql /docker-entrypoint-initdb.d/

COPY initial-notaryserver.sql /docker-entrypoint-initdb.d/ 
COPY initial-notarysigner.sql /docker-entrypoint-initdb.d/