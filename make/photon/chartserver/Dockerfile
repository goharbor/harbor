FROM photon:2.0

COPY ./make/photon/tdnf/photon.internal ./make/photon/tdnf/photon-updates.internal /etc/yum.repos.d/

# use the internal tdnf repo server
ARG USE_INTERNAL_TDNF
RUN if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.repo /etc/yum.repos.d/photon.repo.bk ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.repo /etc/yum.repos.d/photon-updates.repo.bk ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.internal /etc/yum.repos.d/photon.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.internal /etc/yum.repos.d/photon-updates.repo ; fi

RUN tdnf install -y shadow sudo >>/dev/null\
    && tdnf clean all \
    && groupadd -r -g 10000 chart \
    && useradd --no-log-init -m -g 10000 -u 10000 chart

# roll back tdnf repo server
RUN if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.repo.bk /etc/yum.repos.d/photon.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.repo.bk /etc/yum.repos.d/photon-updates.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" != "true" ] ; then rm /etc/yum.repos.d/photon.internal ; fi \
    && if [ "$USE_INTERNAL_TDNF" != "true" ] ; then rm /etc/yum.repos.d/photon-updates.internal ; fi

COPY ./make/photon/chartserver/binary/chartm /home/chart/
COPY ./make/photon/chartserver/docker-entrypoint.sh /home/chart/
COPY ./make/photon/common/install_cert.sh /home/chart/

RUN chown -R chart:chart /etc/pki/tls/certs \
    && chown -R chart:chart /home/chart \
    && chmod u+x /home/chart/chartm \
    && chmod u+x /home/chart/docker-entrypoint.sh \
    && chmod u+x /home/chart/install_cert.sh

USER chart

WORKDIR /home/chart

ENTRYPOINT ["./docker-entrypoint.sh"]

VOLUME ["/chart_storage"]
EXPOSE 9999

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -sS 127.0.0.1:9999/health || exit 1
