from library/photon:1.0

COPY ./binary/notary-signer /bin/notary-signer
COPY ./migrate /bin/migrate
COPY ./migrations/ /migrations/

ENV SERVICE_NAME=notary_signer
ENTRYPOINT [ "notary-signer" ]