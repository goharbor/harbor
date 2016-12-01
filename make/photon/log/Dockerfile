FROM library/photon:1.0

# run logrotate hourly, disable imklog model, provides TCP/UDP syslog reception
RUN tdnf install -y cronie rsyslog shadow tar gzip \
	&& mkdir /etc/rsyslog.d/ \
    && mkdir /var/spool/rsyslog \
    && groupadd syslog \
    && useradd -g syslog syslog
    
ADD make/common/log/rsyslog.conf /etc/rsyslog.conf

# rotate logs weekly
# notes: file name cannot contain dot, or the script will not run
ADD make/common/log/rotate.sh /etc/cron.weekly/rotate

# rsyslog configuration file for docker
ADD make/common/log/rsyslog_docker.conf /etc/rsyslog.d/

VOLUME /var/log/docker/

EXPOSE 514

CMD crond && rm -f /var/run/rsyslogd.pid && rsyslogd -n
