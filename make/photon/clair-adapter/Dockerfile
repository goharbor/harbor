FROM goharbor/harbor-clair-adapter-base:${harbor_base_image_version}

COPY ./make/photon/clair-adapter/binary/harbor-scanner-clair /clair-adapter/clair-adapter

RUN chown -R 10000:10000 /clair-adapter \
    && chmod u+x /clair-adapter/clair-adapter

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -sS 127.0.0.1:8080/probe/healthy || exit 1

USER clair-adapter

ENTRYPOINT ["/clair-adapter/clair-adapter"]