FROM photon:3.0

RUN tdnf install -y sudo >>/dev/null\
    && tdnf clean all \
    && mkdir /clair-adapter/ \
    && groupadd -r -g 10000 clair-adapter \
    && useradd --no-log-init -m -r -g 10000 -u 10000 clair-adapter

COPY ./make/photon/clair-adapter/binary/harbor-scanner-clair /clair-adapter/clair-adapter

RUN chown -R 10000:10000 /clair-adapter \
    && chmod u+x /clair-adapter/clair-adapter

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -sS 127.0.0.1:8080/probe/healthy || exit 1

USER clair-adapter

ENTRYPOINT ["/clair-adapter/clair-adapter"]