FROM testregistry.dataman.io/mysql:5.6

WORKDIR /tmp

ADD registry.sql r.sql 

ADD docker-entrypoint.sh /entrypoint.sh
RUN chmod u+x /entrypoint.sh 
