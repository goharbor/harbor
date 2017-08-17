FROM mysql:5.6.35

VOLUME ["/tmp"]
VOLUME ["/var/run/mysqld"]

WORKDIR /tmp

ADD registry.sql r.sql 

ADD docker-entrypoint.sh /entrypoint.sh
RUN chmod u+x /entrypoint.sh 
