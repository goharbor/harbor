ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-notary-signer-base:${harbor_base_image_version}

ARG TARGETARCH
COPY ./make/photon/notary/binary/migrate-patch-linux-${TARGETARCH:-amd64} /bin/migrate-patch
COPY ./make/photon/notary/binary/notary-signer-linux-${TARGETARCH:-amd64} /bin/notary-signer
COPY ./make/photon/notary/binary/migrate-linux-${TARGETARCH:-amd64} /bin/migrate
COPY ./make/photon/notary/binary/migrations/ /migrations/

RUN chmod +x /bin/notary-signer /migrations/migrate.sh /bin/migrate /bin/migrate-patch
ENV SERVICE_NAME=notary_signer
USER notary
CMD migrate-patch -database=${DB_URL} && /migrations/migrate.sh && /bin/notary-signer -config=/etc/notary/signer-config.postgres.json -logf=logfmt