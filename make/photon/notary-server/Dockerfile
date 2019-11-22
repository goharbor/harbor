ARG harbor_base_image_version
FROM goharbor/harbor-notary-server-base:${harbor_base_image_version}

COPY ./make/photon/notary/migrate-patch /bin/migrate-patch
COPY ./make/photon/notary/binary/notary-server /bin/notary-server
COPY ./make/photon/notary/binary/migrate /bin/migrate
COPY ./make/photon/notary/binary/migrations/ /migrations/

RUN chmod +x /bin/notary-server /migrations/migrate.sh /bin/migrate /bin/migrate-patch
ENV SERVICE_NAME=notary_server
USER notary
CMD migrate-patch -database=${DB_URL} && /migrations/migrate.sh && /bin/notary-server -config=/etc/notary/server-config.postgres.json -logf=logfmt