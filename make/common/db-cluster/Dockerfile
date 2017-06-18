FROM mysql/mysql-cluster:7.6

COPY my.cnf /etc/my.cnf

ADD registry.sql r.sql 

ADD docker-entrypoint.sh /entrypoint.sh
RUN chmod u+x /entrypoint.sh 
