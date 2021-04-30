ARG harbor_base_image_version
ARG harbor_base_namespace

FROM golang:1.15.6 AS golang
RUN git clone "https://github.com/go-delve/delve"
WORKDIR /go/delve
RUN go install "github.com/go-delve/delve/cmd/dlv"

FROM ${harbor_base_namespace}/harbor-core-base:${harbor_base_image_version}

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080/api/v2.0/ping || curl -k --fail -s https://127.0.0.1:8443/api/v2.0/ping || exit 1
COPY ./make/photon/common/install_cert.sh /harbor/
COPY ./make/photon/core/entrypoint.sh /harbor/
COPY ./make/photon/core/harbor_core /harbor/
COPY ./src/core/views /harbor/views
COPY ./make/migrations /harbor/migrations
COPY ./icons /harbor/icons
COPY --from=golang /go/bin/dlv /usr/local/bin/dlv

RUN chown -R harbor:harbor /etc/pki/tls/certs \
    && chown -R harbor:harbor /harbor/ \
    && chmod u+x /harbor/entrypoint.sh \
    && chmod u+x /harbor/install_cert.sh \
    && chmod u+x /harbor/harbor_core

WORKDIR /harbor/
USER harbor
ENTRYPOINT ["/harbor/entrypoint.sh"]
COPY make/photon/prepare/versions /harbor/
EXPOSE 40000
