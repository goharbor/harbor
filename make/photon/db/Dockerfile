FROM photon:2.0

ENV PGDATA /var/lib/postgresql/data

COPY ./make/photon/tdnf/photon.internal ./make/photon/tdnf/photon-updates.internal /etc/yum.repos.d/

# use the internal tdnf repo server
ARG USE_INTERNAL_TDNF
RUN if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.repo /etc/yum.repos.d/photon.repo.bk ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.repo /etc/yum.repos.d/photon-updates.repo.bk ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.internal /etc/yum.repos.d/photon.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.internal /etc/yum.repos.d/photon-updates.repo ; fi

RUN tdnf install -y shadow gzip postgresql >> /dev/null\
    && groupadd -r postgres --gid=999 \
    && useradd -m -r -g postgres --uid=999 postgres \
    && mkdir -p /docker-entrypoint-initdb.d \
    && mkdir -p /run/postgresql \
    && chown -R postgres:postgres /run/postgresql \
    && chmod 2777 /run/postgresql \
    && mkdir -p "$PGDATA" && chown -R postgres:postgres "$PGDATA" && chmod 777 "$PGDATA" \
    && sed -i "s|#listen_addresses = 'localhost'.*|listen_addresses = '*'|g" /usr/share/postgresql/postgresql.conf.sample \
    && sed -i "s|#unix_socket_directories = '/tmp'.*|unix_socket_directories = '/run/postgresql'|g" /usr/share/postgresql/postgresql.conf.sample \
    && tdnf clean all

RUN tdnf erase -y toybox && tdnf install -y util-linux net-tools

# roll back tdnf repo server
RUN if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.repo.bk /etc/yum.repos.d/photon.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.repo.bk /etc/yum.repos.d/photon-updates.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" != "true" ] ; then rm /etc/yum.repos.d/photon.internal ; fi \
    && if [ "$USE_INTERNAL_TDNF" != "true" ] ; then rm /etc/yum.repos.d/photon-updates.internal ; fi

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
