FROM golang:1.10.1-alpine

RUN apk add --update git gcc libc-dev

# Pin to the specific v3.0.0 version
RUN go get -tags 'mysql postgres file' github.com/mattes/migrate/cli && mv /go/bin/cli /go/bin/migrate

ENV NOTARYPKG github.com/theupdateframework/notary

# Copy the local repo to the expected go path
COPY . /go/src/${NOTARYPKG}

WORKDIR /go/src/${NOTARYPKG}

RUN chmod 0600 ./fixtures/database/*

ENV SERVICE_NAME=notary_signer
ENV NOTARY_SIGNER_DEFAULT_ALIAS="timestamp_1"
ENV NOTARY_SIGNER_TIMESTAMP_1="testpassword"

# Install notary-signer
RUN go install \
    -tags pkcs11 \
    -ldflags "-w -X ${NOTARYPKG}/version.GitCommit=`git rev-parse --short HEAD` -X ${NOTARYPKG}/version.NotaryVersion=`cat NOTARY_VERSION`" \
    ${NOTARYPKG}/cmd/notary-signer && apk del git gcc libc-dev && rm -rf /var/cache/apk/*

ENTRYPOINT [ "notary-signer" ]
CMD [ "-config=fixtures/signer-config-local.json" ]
