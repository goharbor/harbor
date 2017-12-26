FROM vmware/mariadb-photon:10.2.8

RUN tdnf distro-sync -y \
    && tdnf install -y mariadb-devel python2 python2-devel python-pip gcc \
    linux-api-headers glibc-devel binutils zlib-devel openssl-devel \
    && pip install mysqlclient alembic \
    && tdnf clean all \
    && mkdir -p /harbor-migration

WORKDIR /harbor-migration

COPY ./ ./

ENTRYPOINT ["./run.sh"] 
