FROM vmware/photon:1.0

RUN tdnf distro-sync -y || echo \
    && tdnf install -y cronie rsyslog shadow tar gzip sudo net-tools\
    && mkdir /etc/rsyslog.d/ \
    && mkdir /var/spool/rsyslog \
    && groupadd -r -g 10000 syslog && useradd --no-log-init -r -g 10000 -u 10000 syslog \
    && tdnf clean all

COPY rsyslog.conf /etc/rsyslog.conf

# notes: file name cannot contain dot, or the script will not run
COPY rotate.sh /etc/cron.daily/rotate

# rsyslog configuration file for docker
COPY rsyslog_docker.conf /etc/rsyslog.d/

COPY start.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/start.sh && \
    chown -R 10000:10000 /run 

HEALTHCHECK CMD netstat -ltu|grep 10514

VOLUME /var/log/docker/ /run/

EXPOSE 10514

CMD /usr/local/bin/start.sh
