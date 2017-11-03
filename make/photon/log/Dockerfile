FROM vmware/rsyslog-photon:8.15.0

COPY rsyslog.conf /etc/rsyslog.conf

# rotate logs weekly
# notes: file name cannot contain dot, or the script will not run
COPY rotate.sh /etc/cron.daily/rotate

# rsyslog configuration file for docker
COPY rsyslog_docker.conf /etc/rsyslog.d/

COPY start.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/start.sh && \
    tdnf install -y sudo net-tools && \
    chown -R 10000:10000 /run 

HEALTHCHECK CMD netstat -ltu|grep 10514

VOLUME /var/log/docker/ /run/

EXPOSE 10514

#CMD crond && rm -f /var/run/rsyslogd.pid && rsyslogd -n
CMD /usr/local/bin/start.sh
