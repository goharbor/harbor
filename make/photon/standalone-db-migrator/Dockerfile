ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-db-base:${harbor_base_image_version}

ENV EXTERNAL_DB 0

COPY --link --chown=999:999 ./make/migrations /migrations
COPY --link --chown=999:999 --chmod=755 ./make/photon/standalone-db-migrator/migrate /harbor/
COPY --link --chown=999:999 --chmod=744 ./make/photon/standalone-db-migrator/entrypoint.sh /harbor/

USER postgres

ENTRYPOINT ["/harbor/entrypoint.sh"]
