FROM photon:2.0

RUN tdnf install -y shadow sudo \
    && tdnf clean all \
    && groupadd -r -g 10000 notary \
    && useradd --no-log-init -r -g 10000 -u 10000 notary
COPY ./make/photon/notary/binary/notary-signer /bin/notary-signer
COPY ./make/photon/notary/binary/migrate /bin/migrate
COPY ./make/photon/notary/binary/migrations/ /migrations/
COPY ./make/photon/notary/signer-start.sh /bin/signer-start.sh

RUN chmod u+x /bin/notary-signer /migrations/migrate.sh /bin/migrate /bin/signer-start.sh
ENV SERVICE_NAME=notary_signer
ENTRYPOINT [ "/bin/signer-start.sh" ]
