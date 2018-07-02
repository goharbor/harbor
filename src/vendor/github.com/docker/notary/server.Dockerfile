FROM golang:1.8.3-alpine
MAINTAINER David Lawrence "david.lawrence@docker.com"

RUN apk add --update git gcc libc-dev && rm -rf /var/cache/apk/*

# Pin to the specific v3.0.0 version
RUN go get -tags 'mysql postgres file' github.com/mattes/migrate/cli && mv /go/bin/cli /go/bin/migrate

ENV NOTARYPKG github.com/docker/notary

# Copy the local repo to the expected go path
COPY . /go/src/${NOTARYPKG}

WORKDIR /go/src/${NOTARYPKG}

RUN chmod 0600 ./fixtures/database/*

ENV SERVICE_NAME=notary_server
EXPOSE 4443

# Install notary-server
RUN go install \
    -tags pkcs11 \
    -ldflags "-w -X ${NOTARYPKG}/version.GitCommit=`git rev-parse --short HEAD` -X ${NOTARYPKG}/version.NotaryVersion=`cat NOTARY_VERSION`" \
    ${NOTARYPKG}/cmd/notary-server && apk del git gcc libc-dev

ENTRYPOINT [ "notary-server" ]
CMD [ "-config=fixtures/server-config-local.json" ]
