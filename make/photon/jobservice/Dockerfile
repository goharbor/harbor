FROM photon:2.0

COPY ./make/photon/tdnf/photon.internal ./make/photon/tdnf/photon-updates.internal /etc/yum.repos.d/

# use the internal tdnf repo server
ARG USE_INTERNAL_TDNF
RUN if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.repo /etc/yum.repos.d/photon.repo.bk ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.repo /etc/yum.repos.d/photon-updates.repo.bk ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.internal /etc/yum.repos.d/photon.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.internal /etc/yum.repos.d/photon-updates.repo ; fi

RUN tdnf install sudo tzdata -y >> /dev/null \
    && tdnf clean all \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -r -g 10000 -u 10000 harbor

# roll back tdnf repo server
RUN if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.repo.bk /etc/yum.repos.d/photon.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.repo.bk /etc/yum.repos.d/photon-updates.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" != "true" ] ; then rm /etc/yum.repos.d/photon.internal ; fi \
    && if [ "$USE_INTERNAL_TDNF" != "true" ] ; then rm /etc/yum.repos.d/photon-updates.internal ; fi

COPY ./make/photon/jobservice/harbor_jobservice /harbor/

RUN chmod u+x /harbor/harbor_jobservice

WORKDIR /harbor/

USER harbor

VOLUME ["/var/log/jobs/"]

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080/api/v1/stats || exit 1

ENTRYPOINT ["/harbor/harbor_jobservice", "-c", "/etc/jobservice/config.yml"]
