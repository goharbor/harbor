FROM library/ubuntu:14.04

# run logrotate hourly, disable imklog model, provides TCP/UDP syslog reception
RUN mv /etc/cron.daily/logrotate /etc/cron.hourly/ \
	&& rm /etc/rsyslog.d/* \
        && rm /etc/rsyslog.conf
ADD make/common/log/rsyslog.conf /etc/rsyslog.conf

# logrotate configuration file for docker
ADD make/common/log/logrotate_docker.conf /etc/logrotate.d/

# rsyslog configuration file for docker
ADD make/common/log/rsyslog_docker.conf /etc/rsyslog.d/

VOLUME /var/log/docker/

EXPOSE 514

CMD cron && rsyslogd -n
