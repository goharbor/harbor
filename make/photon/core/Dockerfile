# syntax=docker/dockerfile:1
# check=skip=InvalidDefaultArgInFrom

ARG harbor_base_image_version
ARG harbor_base_namespace

FROM ${harbor_base_namespace}/harbor-core-base:${harbor_base_image_version}

COPY ./make/photon/common/install_cert.sh /harbor/
COPY ./make/photon/core/entrypoint.sh /harbor/
COPY ./make/photon/core/harbor_core /harbor/
COPY ./src/core/views /harbor/views
COPY ./make/migrations /harbor/migrations
COPY ./icons /harbor/icons

RUN chown -R harbor:harbor /etc/pki/tls/certs \
    && chown -R harbor:harbor /harbor/ \
    && chmod u+x /harbor/entrypoint.sh \
    && chmod u+x /harbor/install_cert.sh \
    && chmod u+x /harbor/harbor_core

WORKDIR /harbor/
USER harbor

COPY make/photon/prepare/versions /harbor/

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl --fail -s http://localhost:8080/api/v2.0/ping || curl -k --fail -s https://localhost:8443/api/v2.0/ping || exit 1

ENTRYPOINT ["/harbor/entrypoint.sh"]
