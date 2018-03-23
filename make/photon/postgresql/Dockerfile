FROM vmware/photon:1.0

ENV PGDATA /var/lib/postgresql/data

RUN touch /etc/localtime.bak \
    && tdnf distro-sync -y \
    && tdnf install -y sed shadow gzip postgresql >> /dev/null\
    && groupadd -r postgres --gid=999 \
    && useradd -r -g postgres --uid=999 postgres \
    && mkdir -p /docker-entrypoint-initdb.d \
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

ADD docker-entrypoint.sh /entrypoint.sh
RUN chmod u+x /entrypoint.sh
ENTRYPOINT ["/entrypoint.sh"]

EXPOSE 5432
CMD ["postgres"]
