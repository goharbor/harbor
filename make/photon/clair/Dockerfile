FROM photon:2.0

COPY ./make/photon/tdnf/photon.internal ./make/photon/tdnf/photon-updates.internal /etc/yum.repos.d/

# use the internal tdnf repo server
ARG USE_INTERNAL_TDNF
RUN if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.repo /etc/yum.repos.d/photon.repo.bk ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.repo /etc/yum.repos.d/photon-updates.repo.bk ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.internal /etc/yum.repos.d/photon.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.internal /etc/yum.repos.d/photon-updates.repo ; fi

RUN tdnf install -y git shadow sudo rpm xz python-xml >>/dev/null\
    && tdnf clean all \
    && groupadd -r -g 10000 clair \
    && useradd --no-log-init -m -g 10000 -u 10000 clair

# roll back tdnf repo server
RUN if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.repo.bk /etc/yum.repos.d/photon.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.repo.bk /etc/yum.repos.d/photon-updates.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" != "true" ] ; then rm /etc/yum.repos.d/photon.internal ; fi \
    && if [ "$USE_INTERNAL_TDNF" != "true" ] ; then rm /etc/yum.repos.d/photon-updates.internal ; fi

COPY ./make/photon/clair/binary/clair /home/clair/
COPY ./make/photon/clair/docker-entrypoint.sh /home/clair/
COPY ./make/photon/clair/dumb-init /home/clair/
COPY ./make/photon/common/install_cert.sh /home/clair/

VOLUME /config

EXPOSE 6060 6061

RUN chown -R clair:clair /etc/pki/tls/certs \
    && chown -R clair:clair /home/clair \
    && chmod u+x /home/clair/clair \
    && chmod u+x /home/clair/docker-entrypoint.sh \
    && chmod +x /home/clair/dumb-init

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -sS 127.0.0.1:6061/health || exit 1

WORKDIR /home/clair
USER clair
ENTRYPOINT ["./docker-entrypoint.sh"]
