FROM vmware/photon:1.0

RUN tdnf distro-sync -y \
    && tdnf erase vim -y \
    && tdnf clean all
COPY ./binary/notary-server /bin/notary-server
COPY ./migrate /bin/migrate
COPY ./migrations/ /migrations/

ENV SERVICE_NAME=notary_server
ENTRYPOINT [ "notary-server" ]
