FROM cicd.harbor.bitsf.xin/harbor-depend/chartserver:base

COPY ./make/photon/chartserver/binary/chartm /chartserver/
COPY ./make/photon/chartserver/docker-entrypoint.sh /docker-entrypoint.sh
COPY ./make/photon/common/install_cert.sh /harbor

VOLUME ["/chart_storage"]
EXPOSE 9999

RUN chown -R 10000:10000 /chartserver \
    && chmod u+x /chartserver/chartm \
    && chmod u+x /docker-entrypoint.sh

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -sS 127.0.0.1:9999/health || exit 1

ENTRYPOINT ["/docker-entrypoint.sh"]
