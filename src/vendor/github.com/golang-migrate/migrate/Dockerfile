FROM golang:1.11-alpine3.8 AS downloader
ARG VERSION

RUN apk add --no-cache git gcc musl-dev curl

RUN curl -fsSL -o /usr/local/bin/dep https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 && chmod +x /usr/local/bin/dep

WORKDIR /go/src/github.com/golang-migrate/migrate

COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only

COPY *.go ./
COPY cli ./cli
COPY database ./database
COPY source ./source

ENV DATABASES="postgres mysql redshift cassandra spanner cockroachdb clickhouse"
ENV SOURCES="file go_bindata github aws_s3 google_cloud_storage"

RUN go build -a -o build/migrate.linux-386 -ldflags="-X main.Version=${VERSION}" -tags "$DATABASES $SOURCES" ./cli

FROM alpine:3.8

RUN apk add --no-cache ca-certificates

COPY --from=downloader /go/src/github.com/golang-migrate/migrate/build/migrate.linux-386 /migrate

ENTRYPOINT ["/migrate"]
CMD ["--help"]
