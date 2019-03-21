FROM photon:2.0

ENV PGDATA /var/lib/postgresql/data

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

VOLUME /var/lib/postgresql/data

ADD ./make/photon/db/docker-entrypoint.sh /entrypoint.sh
ADD ./make/photon/db/docker-healthcheck.sh /docker-healthcheck.sh
RUN chmod u+x /entrypoint.sh /docker-healthcheck.sh
ENTRYPOINT ["/entrypoint.sh"]
HEALTHCHECK CMD ["/docker-healthcheck.sh"]

COPY ./make/photon/db/initial-notaryserver.sql /docker-entrypoint-initdb.d/
COPY ./make/photon/db/initial-notarysigner.sql /docker-entrypoint-initdb.d/
COPY ./make/photon/db/initial-registry.sql /docker-entrypoint-initdb.d/

EXPOSE 5432
CMD ["postgres"]
