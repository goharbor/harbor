FROM photon:2.0

ENV PGDATA /var/lib/postgresql/data

## have both mysql and pgsql installed.
RUN tdnf distro-sync -y \
    && tdnf remove -y toybox \
    && tdnf install -y sed shadow procps-ng gawk gzip sudo net-tools glibc-i18n >> /dev/null\
    && tdnf install -y python2 python2-devel python-pip gcc PyYAML python-jinja2 \
    linux-api-headers glibc-devel binutils zlib-devel openssl-devel postgresql python-psycopg2 >> /dev/null \
    && pip install alembic \
    && mkdir /docker-entrypoint-initdb.d /docker-entrypoint-updatedb.d \
    && chmod 777 /docker-entrypoint-initdb.d /docker-entrypoint-updatedb.d \
    && mkdir -p /harbor-migration \
    && touch /etc/localtime.bak \
    && groupadd -r postgres --gid=999 \
    && useradd -r -g postgres --uid=999 postgres \
    && mkdir -p /run/postgresql \
    && chown -R postgres:postgres /run/postgresql \
    && chmod 2777 /run/postgresql \
    && mkdir -p "$PGDATA" && chown -R postgres:postgres "$PGDATA" && chmod 777 "$PGDATA" \
    && sed -i "s|#listen_addresses = 'localhost'.*|listen_addresses = '*'|g" /usr/share/postgresql/postgresql.conf.sample \
    && sed -i "s|#unix_socket_directories = '/tmp'.*|unix_socket_directories = '/run/postgresql'|g" /usr/share/postgresql/postgresql.conf.sample \
    && touch /usr/share/locale/locale.alias \
    && locale-gen.sh en_US.UTF-8 \
    && tdnf clean all

VOLUME /var/lib/postgresql/data

WORKDIR /harbor-migration

COPY ./ ./

ENTRYPOINT ["./docker-entrypoint.sh"]
