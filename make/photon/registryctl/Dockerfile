ARG harbor_base_image_version
ARG harbor_base_namespace
ARG GOBUILDIMAGE
FROM ${GOBUILDIMAGE} AS builder
WORKDIR /harbor
COPY ../../ /harbor
ENV CGO_ENABLED=0
RUN apt update \
    && apt install -y make \
    && cd /harbor \
    && make registryctl

ARG REGISTRY_SRC_TAG
ENV DISTRIBUTION_DIR=/go/src/github.com/docker
ENV BUILDTAGS="include_oss include_gcs"
ENV GO111MODULE=auto


WORKDIR $DISTRIBUTION_DIR

RUN  git clone -b $REGISTRY_SRC_TAG https://github.com/distribution/distribution.git  \
     && cd distribution \
     && git apply /harbor/make/photon/registry/redis.patch \
     && mkdir -p binary \
     && rm -rf binary/registry || true \
     && make PREFIX=/go clean binaries



FROM ${harbor_base_namespace}/harbor-registryctl-base:${harbor_base_image_version}

COPY --from=builder /harbor/make/photon/common/install_cert.sh /home/harbor
COPY --from=builder /go/src/github.com/docker/distribution/bin/registry /usr/bin/registry_DO_NOT_USE_GC
COPY --from=builder /harbor/make/photon/registryctl/start.sh /home/harbor
COPY --from=builder /harbor/make/photon/registryctl/harbor_registryctl /home/harbor

RUN chown harbor:harbor /home/harbor/harbor_registryctl && chmod u+x /home/harbor/harbor_registryctl \
    && chown harbor:harbor /usr/bin/registry_DO_NOT_USE_GC && chmod u+x /usr/bin/registry_DO_NOT_USE_GC \
    && chown harbor:harbor /home/harbor/start.sh && chmod u+x /home/harbor/start.sh \
    && chown harbor:harbor /home/harbor/install_cert.sh && chmod u+x /home/harbor/install_cert.sh


HEALTHCHECK CMD curl --fail -s http://localhost:8080/api/health || curl -sk --fail --key /etc/harbor/ssl/registryctl.key --cert /etc/harbor/ssl/registryctl.crt https://localhost:8443/api/health || exit 1

VOLUME ["/var/lib/registry"]

ENTRYPOINT ["/home/harbor/start.sh"]

USER harbor
