FROM vmware/photon:1.0


RUN tdnf distro-sync -y \
    && tdnf erase vim -y \
    && tdnf install -y git shadow sudo bzr rpm xz python-xml \
    && tdnf clean all \
    && mkdir /clair2.0.1/ \
    && groupadd -r -g 10000 clair \
    && useradd --no-log-init -m -r -g 10000 -u 10000 clair
COPY clair /clair2.0.1/
COPY docker-entrypoint.sh /docker-entrypoint.sh
COPY dumb-init /dumb-init 

VOLUME /config

EXPOSE 6060 6061

RUN chown -R 10000:10000 /clair2.0.1 \
    && chmod u+x /clair2.0.1/clair \
    && chmod u+x /docker-entrypoint.sh \
    && chmod +x /dumb-init


HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -sS 127.0.0.1:6061/health || exit 1

USER clair

ENTRYPOINT ["/docker-entrypoint.sh"]
