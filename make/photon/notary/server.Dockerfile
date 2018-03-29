FROM vmware/photon:1.0

RUN tdnf distro-sync -y \
    && tdnf erase vim -y \
    && tdnf install -y shadow sudo \
    && tdnf clean all \
    && groupadd -r -g 10000 notary \
    && useradd --no-log-init -r -g 10000 -u 10000 notary

COPY ./binary/notary-server /bin/notary-server
COPY ./binary/migrate /bin/migrate
COPY ./binary/migrations/ /migrations/
COPY ./server-start.sh /bin/server-start.sh
RUN chmod u+x /bin/notary-server /migrations/migrate.sh /bin/migrate /bin/server-start.sh
ENV SERVICE_NAME=notary_server
ENTRYPOINT [ "/bin/server-start.sh" ]