ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-core-base:${harbor_base_image_version}

HEALTHCHECK CMD curl --fail -s http://localhost:8080/api/v2.0/ping || curl -k --fail -s https://localhost:8443/api/v2.0/ping || exit 1
COPY --link --chown=10000:10000 --chmod=755 ./make/photon/common/install_cert.sh /harbor/
COPY --link --chown=10000:10000 --chmod=744 ./make/photon/core/entrypoint.sh /harbor/
COPY --link --chown=10000:10000 --chmod=755 ./make/photon/core/harbor_core /harbor/
COPY --link --chown=10000:10000 ./src/core/views /harbor/views
COPY --link --chown=10000:10000 ./make/migrations /harbor/migrations
COPY --link --chown=10000:10000 ./icons /harbor/icons

WORKDIR /harbor/
USER harbor
ENTRYPOINT ["/harbor/entrypoint.sh"]
COPY --link make/photon/prepare/versions /harbor/
