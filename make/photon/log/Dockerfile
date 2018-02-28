FROM vmware/photon:1.0

RUN tdnf distro-sync -y \
    && tdnf install -y cronie rsyslog logrotate shadow tar gzip sudo net-tools >> /dev/null\
    && mkdir /etc/rsyslog.d/ \
    && mkdir /var/spool/rsyslog \
    && groupadd -r -g 10000 syslog && useradd --no-log-init -r -g 10000 -u 10000 syslog \
    && tdnf clean all

COPY rsyslog.conf /etc/rsyslog.conf

# rsyslog configuration file for docker
COPY rsyslog_docker.conf /etc/rsyslog.d/

# run logrotate hourly
RUN mv /etc/cron.daily/logrotate /etc/cron.hourly/logrotate

COPY start.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/start.sh && \
    chown -R 10000:10000 /etc/rsyslog.conf /etc/rsyslog.d/ /run 

HEALTHCHECK CMD netstat -ltu|grep 10514

VOLUME /var/log/docker/ /run/ /etc/logrotate.d/

EXPOSE 10514

CMD /usr/local/bin/start.sh
