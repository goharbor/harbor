FROM photon:4.0

ENV PGDATA /var/lib/postgresql/data

COPY ./make/photon/db/postgresql96-libs-9.6.21-1.ph4.x86_64.rpm /pg96/
COPY ./make/photon/db/postgresql96-9.6.21-1.ph4.x86_64.rpm /pg96/

RUN tdnf install -y /pg96/postgresql96-libs-9.6.21-1.ph4.x86_64.rpm /pg96/postgresql96-9.6.21-1.ph4.x86_64.rpm >> /dev/null \
    && rm -rf /pg96 \
    && tdnf install -y shadow gzip postgresql13 findutils bc >> /dev/null \
    && groupadd -r postgres --gid=999 \
    && useradd -m -r -g postgres --uid=999 postgres \
    && mkdir -p /docker-entrypoint-initdb.d \
    && mkdir -p /run/postgresql \
    && chown -R postgres:postgres /run/postgresql \
    && chmod 2777 /run/postgresql \
    && mkdir -p "$PGDATA" && chown -R postgres:postgres "$PGDATA" && chmod 777 "$PGDATA" \
    && sed -i "s|#listen_addresses = 'localhost'.*|listen_addresses = '*'|g" /usr/pgsql/13/share/postgresql.conf.sample \
    && sed -i "s|#unix_socket_directories = '/tmp'.*|unix_socket_directories = '/run/postgresql'|g" /usr/pgsql/13/share/postgresql.conf.sample \
    && ln -s /usr/pgsql/13/bin/* /usr/bin/ \
    && tdnf clean all

RUN tdnf erase -y toybox && tdnf install -y util-linux net-tools