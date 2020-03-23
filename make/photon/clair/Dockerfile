ARG harbor_base_image_version
FROM goharbor/harbor-clair-base:${harbor_base_image_version}

COPY ./make/photon/clair/binary/clair /home/clair/
COPY ./make/photon/clair/docker-entrypoint.sh /home/clair/
COPY ./make/photon/clair/dumb-init /home/clair/
COPY ./make/photon/common/install_cert.sh /home/clair/

VOLUME /config

EXPOSE 6060 6061

RUN chown -R clair:clair /etc/pki/tls/certs && chown -R clair:clair /home/clair \
    && chmod u+x /home/clair/clair \
    && chmod u+x /home/clair/docker-entrypoint.sh \
    && chmod u+x /home/clair/install_cert.sh \
    && chmod +x /home/clair/dumb-init

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -sS 127.0.0.1:6061/health || exit 1

WORKDIR /home/clair
USER clair
ENTRYPOINT ["./docker-entrypoint.sh"]
