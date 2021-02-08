FROM mysql:5.7

RUN groupadd -r mysql --gid=999 \
    && useradd -m -r -g mysql --uid=999 mysql

VOLUME /var/lib/mysql

COPY ./make/photon/db-mysql/initial-notaryserver.sql /docker-entrypoint-initdb.d/
COPY ./make/photon/db-mysql/initial-notarysigner.sql /docker-entrypoint-initdb.d/
COPY ./make/photon/db-mysql/initial-registry.sql /docker-entrypoint-initdb.d/

ENTRYPOINT ["/entrypoint.sh"]
HEALTHCHECK CMD /healthcheck.sh
EXPOSE 3306 33060
CMD ["mysqld"]

USER mysql
