ARG harbor_base_image_version
FROM goharbor/harbor-core-base:${harbor_base_image_version}

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080/api/ping || exit 1
COPY ./make/photon/core/harbor_core /harbor/
COPY ./src/core/views /harbor/views
COPY ./make/migrations /harbor/migrations

RUN chmod u+x /harbor/harbor_core
WORKDIR /harbor/
USER harbor
ENTRYPOINT ["/harbor/harbor_core"]
