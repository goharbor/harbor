from library/photon:1.0

COPY ./binary/notary-server /bin/notary-server
COPY ./migrate /bin/migrate
COPY ./migrations/ /migrations/

ENV SERVICE_NAME=notary_server
ENTRYPOINT [ "notary-server" ]