FROM vmware/photon:1.0

RUN tdnf distro-sync -y \
    && tdnf erase vim -y \
    && tdnf clean all
COPY ./binary/notary-signer /bin/notary-signer
COPY ./migrate /bin/migrate
COPY ./migrations/ /migrations/

ENV SERVICE_NAME=notary_signer
ENTRYPOINT [ "notary-signer" ]
