FROM vmware/photon:1.0

#The Docker Daemon has to be running with storage backend btrfs when building the image

RUN tdnf distro-sync -y \
    && tdnf install -y sed shadow procps-ng gawk gzip sudo net-tools >> /dev/null\
    && groupadd -r -g 10000 mysql && useradd --no-log-init -r -g 10000 -u 10000 mysql \
    && tdnf install -y mariadb-server mariadb >> /dev/null\
    && mkdir /docker-entrypoint-initdb.d /docker-entrypoint-updatedb.d \
    && rm -fr /var/lib/mysql \
    && mkdir -p /var/lib/mysql /var/run/mysqld \
    && chown -R mysql:mysql /var/lib/mysql /var/run/mysqld \
    && chmod 777 /var/run/mysqld /docker-entrypoint-initdb.d /docker-entrypoint-updatedb.d \
    && tdnf clean all

COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh
COPY my.cnf /etc/
RUN ln -s usr/local/bin/docker-entrypoint.sh /

VOLUME /var/lib/mysql /docker-entrypoint-initdb.d /docker-entrypoint-updatedb.d /tmp /var/run/mysqld
EXPOSE 3306

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["mysqld"]
