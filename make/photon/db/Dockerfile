ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-db-base:${harbor_base_image_version}

VOLUME /var/lib/postgresql/data

COPY ./make/photon/db/docker-entrypoint.sh /docker-entrypoint.sh
COPY ./make/photon/db/docker-healthcheck.sh /docker-healthcheck.sh
COPY ./make/photon/db/initial-notaryserver.sql /docker-entrypoint-initdb.d/
COPY ./make/photon/db/initial-notarysigner.sql /docker-entrypoint-initdb.d/
COPY ./make/photon/db/initial-registry.sql /docker-entrypoint-initdb.d/
RUN chown -R postgres:postgres /docker-entrypoint.sh /docker-healthcheck.sh /docker-entrypoint-initdb.d \
    && chmod u+x /docker-entrypoint.sh /docker-healthcheck.sh

ENTRYPOINT ["/docker-entrypoint.sh"]
HEALTHCHECK CMD ["/docker-healthcheck.sh"]

EXPOSE 5432
USER postgres
