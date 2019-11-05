FROM photon:2.0

COPY ./make/photon/tdnf/photon.internal ./make/photon/tdnf/photon-updates.internal /etc/yum.repos.d/

# use the internal tdnf repo server
ARG USE_INTERNAL_TDNF
RUN if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.repo /etc/yum.repos.d/photon.repo.bk ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.repo /etc/yum.repos.d/photon-updates.repo.bk ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.internal /etc/yum.repos.d/photon.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.internal /etc/yum.repos.d/photon-updates.repo ; fi

RUN tdnf install -y sudo >>/dev/null\
    && tdnf clean all \
    && mkdir /clair-adapter/ \
    && groupadd -r -g 10000 clair-adapter \
    && useradd --no-log-init -m -r -g 10000 -u 10000 clair-adapter

# roll back tdnf repo server
RUN if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.repo.bk /etc/yum.repos.d/photon.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.repo.bk /etc/yum.repos.d/photon-updates.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" != "true" ] ; then rm /etc/yum.repos.d/photon.internal ; fi \
    && if [ "$USE_INTERNAL_TDNF" != "true" ] ; then rm /etc/yum.repos.d/photon-updates.internal ; fi

COPY ./make/photon/clair-adapter/binary/harbor-scanner-clair /clair-adapter/clair-adapter

RUN chown -R 10000:10000 /clair-adapter \
    && chmod u+x /clair-adapter/clair-adapter

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -sS 127.0.0.1:8080/probe/healthy || exit 1

USER clair-adapter

ENTRYPOINT ["/clair-adapter/clair-adapter"]