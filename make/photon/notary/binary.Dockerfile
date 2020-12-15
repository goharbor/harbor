FROM golang:1.15.6

ARG NOTARY_VERSION
ARG MIGRATE_VERSION
RUN test -n "$NOTARY_VERSION"
RUN test -n "$MIGRATE_VERSION"
ENV NOTARYPKG github.com/theupdateframework/notary
ENV MIGRATEPKG github.com/golang-migrate/migrate

RUN git clone -b $NOTARY_VERSION https://github.com/theupdateframework/notary.git /go/src/${NOTARYPKG}
WORKDIR /go/src/${NOTARYPKG}

ARG TARGETARCHS=amd64

RUN set -eux; \
    \
    for arch in ${TARGETARCHS}; do \
        GOARCH=${arch} go build -i -o /go/bin/notary-server-linux-${arch} -tags pkcs11 \
            -ldflags "-w -X ${NOTARYPKG}/version.GitCommit=`git rev-parse --short HEAD` -X ${NOTARYPKG}/version.NotaryVersion=`cat NOTARY_VERSION`" ${NOTARYPKG}/cmd/notary-server; \
        GOARCH=${arch} go build -i -o /go/bin/notary-signer-linux-${arch} -tags pkcs11 \
            -ldflags "-w -X ${NOTARYPKG}/version.GitCommit=`git rev-parse --short HEAD` -X ${NOTARYPKG}/version.NotaryVersion=`cat NOTARY_VERSION`" ${NOTARYPKG}/cmd/notary-signer; \
    done

RUN cp -r /go/src/${NOTARYPKG}/migrations/ /

RUN set -eux; git clone -b $MIGRATE_VERSION https://github.com/golang-migrate/migrate /go/src/${MIGRATEPKG}

WORKDIR /go/src/${MIGRATEPKG}

ENV DATABASES="postgres mysql redshift cassandra spanner cockroachdb"
ENV SOURCES="file go_bindata github aws_s3 google_cloud_storage"

RUN set -eux; \
    \
    for arch in ${TARGETARCHS}; do \
        GOARCH=${arch} go build -i -o /go/bin/migrate-linux-${arch} -tags "$DATABASES $SOURCES" -ldflags="-X main.Version=${MIGRATE_VERSION}" ${MIGRATEPKG}/cli; \
    done
