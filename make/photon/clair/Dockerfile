FROM cicd.harbor.bitsf.xin/harbor-depend/clair:base

COPY ./make/photon/clair/binary/clair /clair/
COPY ./make/photon/clair/docker-entrypoint.sh /docker-entrypoint.sh
COPY ./make/photon/clair/dumb-init /dumb-init
COPY ./make/photon/common/install_cert.sh /harbor 

VOLUME /config

EXPOSE 6060 6061

RUN chown -R 10000:10000 /clair \
    && chmod u+x /clair/clair \
    && chmod u+x /docker-entrypoint.sh \
    && chmod +x /dumb-init

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -sS 127.0.0.1:6061/health || exit 1

ENTRYPOINT ["/docker-entrypoint.sh"]
