ARG build_image
ARG harbor_base_image_version
ARG harbor_base_namespace

FROM ${build_image} AS build

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY src /harbor/src
WORKDIR /harbor/src/cmd/exporter
RUN go build -o /out/harbor_exporter

FROM ${harbor_base_namespace}/harbor-exporter-base:${harbor_base_image_version}

COPY --from=build /out/harbor_exporter /harbor/harbor_exporter
COPY ./make/photon/exporter/entrypoint.sh ./make/photon/common/install_cert.sh /harbor/

RUN chown -R harbor:harbor /etc/pki/tls/certs \
    && chown -R harbor:harbor /harbor/ \
    && chmod u+x /harbor/entrypoint.sh \
    && chmod u+x /harbor/install_cert.sh \
    && chmod u+x /harbor/harbor_exporter

WORKDIR /harbor
USER harbor

ENTRYPOINT ["/harbor/entrypoint.sh"]
