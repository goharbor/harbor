ARG GOBUILDIMAGE
ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${GOBUILDIMAGE} AS builder
WORKDIR /harbor
COPY ../../ /harbor
ENV CGO_ENABLED=0
RUN apt update \
    && apt install -y make \
    && cd /harbor \
    && make core


ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-core-base:${harbor_base_image_version}

HEALTHCHECK CMD curl --fail -s http://localhost:8080/api/v2.0/ping || curl -k --fail -s https://localhost:8443/api/v2.0/ping || exit 1
COPY --from=builder /harbor/make/photon/common/install_cert.sh /harbor/
COPY --from=builder /harbor/make/photon/core/entrypoint.sh /harbor/
COPY --from=builder /harbor/make/photon/core/harbor_core /harbor/
COPY --from=builder /harbor/src/core/views /harbor/views
COPY --from=builder /harbor/make/migrations /harbor/migrations
COPY --from=builder /harbor/icons /harbor/icons

RUN chown -R harbor:harbor /harbor/ \
    && chmod u+x /harbor/entrypoint.sh \
    && chmod u+x /harbor/install_cert.sh \
    && chmod u+x /harbor/harbor_core

WORKDIR /harbor/
USER harbor
ENTRYPOINT ["/harbor/entrypoint.sh"]
COPY --from=builder /harbor/make/photon/prepare/versions /harbor/
