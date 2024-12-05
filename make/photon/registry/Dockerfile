ARG harbor_base_image_version
ARG harbor_base_namespace
ARG GOBUILDIMAGE
FROM ${GOBUILDIMAGE} AS builder

ARG REGISTRY_SRC_TAG
ENV DISTRIBUTION_DIR=/go/src/github.com/docker
ENV BUILDTAGS="include_oss include_gcs"
ENV GO111MODULE=auto
RUN mkdir -p $DISTRIBUTION_DIR
COPY ./ /harbor

WORKDIR $DISTRIBUTION_DIR
ENV CGO_ENABLED=0
RUN  git clone -b $REGISTRY_SRC_TAG https://github.com/distribution/distribution.git  \
     && cd distribution \
     && git apply /harbor/make/photon/registry/redis.patch \
     && mkdir -p binary \
     && make PREFIX=/go clean binaries



FROM ${harbor_base_namespace}/harbor-registry-base:${harbor_base_image_version}

COPY ./make/photon/common/install_cert.sh /home/harbor
COPY ./make/photon/registry/entrypoint.sh /home/harbor
COPY --from=builder /go/src/github.com/docker/distribution/bin/registry /usr/bin/registry_DO_NOT_USE_GC
COPY --from=builder /go/src/github.com/docker/distribution/cmd/registry/config-example.yml /etc/registry/config.yml

RUN chown harbor:harbor /home/harbor/entrypoint.sh && chmod u+x /home/harbor/entrypoint.sh \
    && chown harbor:harbor /home/harbor/install_cert.sh && chmod u+x /home/harbor/install_cert.sh \
    && chown harbor:harbor /usr/bin/registry_DO_NOT_USE_GC && chmod u+x /usr/bin/registry_DO_NOT_USE_GC

HEALTHCHECK CMD curl --fail -s http://localhost:5000 || curl -k --fail -s https://localhost:5443 || exit 1

USER harbor

ENTRYPOINT ["/home/harbor/entrypoint.sh"]

VOLUME ["/storage"]
