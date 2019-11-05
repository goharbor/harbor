FROM photon:2.0

COPY ./make/photon/tdnf/photon.internal ./make/photon/tdnf/photon-updates.internal /etc/yum.repos.d/

# use the internal tdnf repo server
ARG USE_INTERNAL_TDNF
RUN if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.repo /etc/yum.repos.d/photon.repo.bk ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.repo /etc/yum.repos.d/photon-updates.repo.bk ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.internal /etc/yum.repos.d/photon.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.internal /etc/yum.repos.d/photon-updates.repo ; fi

RUN tdnf install -y cronie rsyslog logrotate shadow tar gzip sudo >> /dev/null\
    && mkdir /var/spool/rsyslog \
    && groupadd -r -g 10000 syslog && useradd --no-log-init -r -g 10000 -u 10000 syslog \
    && tdnf clean all

# roll back tdnf repo server
RUN if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon.repo.bk /etc/yum.repos.d/photon.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" = "true" ] ; then mv /etc/yum.repos.d/photon-updates.repo.bk /etc/yum.repos.d/photon-updates.repo ; fi \
    && if [ "$USE_INTERNAL_TDNF" != "true" ] ; then rm /etc/yum.repos.d/photon.internal ; fi \
    && if [ "$USE_INTERNAL_TDNF" != "true" ] ; then rm /etc/yum.repos.d/photon-updates.internal ; fi

COPY ./make/photon/log/rsyslog.conf /etc/rsyslog.conf

# rsyslog configuration file for docker
COPY ./make/photon/log/rsyslog_docker.conf /etc/rsyslog.d/

# remove the original "logrotate" in directory "/etc/cron.daily/"
# and copy the customized one to directory "/etc/cron.hourly/" 
# to run logrotate hourly
RUN rm /etc/cron.daily/logrotate
COPY ./make/photon/log/logrotate /etc/cron.hourly/

COPY ./make/photon/log/start.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/start.sh /etc/rsyslog.d/ && \
    chown -R 10000:10000 /etc/rsyslog.conf /etc/rsyslog.d/ /run /var/lib/logrotate/

HEALTHCHECK CMD netstat -ltun|grep 10514

VOLUME /var/log/docker/ /run/ /etc/logrotate.d/

EXPOSE 10514

CMD /usr/local/bin/start.sh
