ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-db-base:${harbor_base_image_version}

VOLUME /var/lib/postgresql/data

COPY --link --chown=999:999 --chmod=744 ./make/photon/db/docker-entrypoint.sh /docker-entrypoint.sh
COPY --link ./make/photon/db/initdb.sh /initdb.sh
COPY --link ./make/photon/db/upgrade.sh /upgrade.sh
COPY --link --chown=999:999 --chmod=744 ./make/photon/db/docker-healthcheck.sh /docker-healthcheck.sh
COPY --link --chown=999:999 ./make/photon/db/initial-registry.sql /docker-entrypoint-initdb.d/

ENTRYPOINT ["/docker-entrypoint.sh", "14", "15"]
HEALTHCHECK CMD ["/docker-healthcheck.sh"]

USER postgres
