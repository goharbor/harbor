FROM photon:2.0

RUN tdnf install -y git shadow sudo rpm xz python-xml >>/dev/null\
    && tdnf clean all \
    && groupadd -r -g 10000 clair \
    && useradd --no-log-init -m -g 10000 -u 10000 clair
COPY ./make/photon/clair/binary/clair /home/clair/
COPY ./make/photon/clair/docker-entrypoint.sh /home/clair/
COPY ./make/photon/clair/dumb-init /home/clair/
COPY ./make/photon/common/install_cert.sh /home/clair/

VOLUME /config

EXPOSE 6060 6061

RUN chmod -R 777 /etc/pki/tls/certs \
    && chown -R clair:clair /home/clair \
    && chmod u+x /home/clair/clair \
    && chmod u+x /home/clair/docker-entrypoint.sh \
    && chmod +x /home/clair/dumb-init

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -sS 127.0.0.1:6061/health || exit 1

WORKDIR /home/clair
USER clair
ENTRYPOINT ["./docker-entrypoint.sh"]
