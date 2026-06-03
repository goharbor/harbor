ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-db-base:${harbor_base_image_version}

VOLUME /var/lib/postgresql/data

COPY ./make/photon/db/docker-entrypoint.sh /docker-entrypoint.sh
COPY ./make/photon/db/initdb.sh /initdb.sh
COPY ./make/photon/db/upgrade.sh /upgrade.sh
COPY ./make/photon/db/docker-healthcheck.sh /docker-healthcheck.sh
COPY ./make/photon/db/initial-registry.sql /docker-entrypoint-initdb.d/
RUN chown -R postgres:postgres /docker-entrypoint.sh /initdb.sh /upgrade.sh \
    /docker-healthcheck.sh /docker-entrypoint-initdb.d \
    && chmod u+x /initdb.sh /upgrade.sh /docker-entrypoint.sh /docker-healthcheck.sh

# Exec-form ENTRYPOINT can't expand the per-arch PG major that harbor-db-base
# records in /etc/harbor-pg-new-version, so a shim injects it (15 = upgrade-from
# version). See goharbor/harbor#23317.
RUN printf '#!/bin/bash\nexec /docker-entrypoint.sh 15 "$(cat /etc/harbor-pg-new-version 2>/dev/null || echo 18)" "$@"\n' \
        > /docker-entrypoint-shim.sh \
    && chmod u+x /docker-entrypoint-shim.sh \
    && chown postgres:postgres /docker-entrypoint-shim.sh

ENTRYPOINT ["/docker-entrypoint-shim.sh"]
HEALTHCHECK CMD ["/docker-healthcheck.sh"]

USER postgres
