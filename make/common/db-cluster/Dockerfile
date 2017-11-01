FROM mysql/mysql-cluster:7.6

ADD registry.sql r.sql 

ADD docker-entrypoint.sh /entrypoint.sh
RUN chmod u+x /entrypoint.sh 
