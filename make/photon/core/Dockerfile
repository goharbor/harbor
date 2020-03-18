ARG harbor_base_image_version
FROM goharbor/harbor-core-base:${harbor_base_image_version}

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080/api/v2.0/ping || exit 1
COPY ./make/photon/common/install_cert.sh /harbor/
COPY ./make/photon/core/entrypoint.sh /harbor/
COPY ./make/photon/core/harbor_core /harbor/
COPY ./src/core/views /harbor/views
COPY ./make/migrations /harbor/migrations

RUN chown -R harbor:harbor /etc/pki/tls/certs \
    && chown harbor:harbor /harbor/entrypoint.sh && chmod u+x /harbor/entrypoint.sh \
    && chown harbor:harbor /harbor/install_cert.sh && chmod u+x /harbor/install_cert.sh \
    && chown harbor:harbor /harbor/harbor_core && chmod u+x /harbor/harbor_core

WORKDIR /harbor/
USER harbor
ENTRYPOINT ["/harbor/entrypoint.sh"]
COPY make/photon/prepare/versions /harbor/
