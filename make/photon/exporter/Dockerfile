ARG build_image
FROM ${build_image} AS build

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY src /harbor/src
WORKDIR /harbor/src/cmd/exporter
RUN go build -o /out/harbor_exporter

FROM photon:2.0

RUN tdnf install -y tzdata shadow >> /dev/null \
    && tdnf clean all \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -r -m -g 10000 -u 10000 harbor \
    && mkdir /harbor/

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
