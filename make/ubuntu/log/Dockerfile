FROM library/ubuntu:14.04

RUN rm /etc/rsyslog.d/* && rm /etc/rsyslog.conf

ADD make/common/log/rsyslog.conf /etc/rsyslog.conf

# rotate logs weekly
# notes: file name cannot contain dot, or the script will not run
ADD make/common/log/rotate.sh /etc/cron.weekly/rotate

# rsyslog configuration file for docker
ADD make/common/log/rsyslog_docker.conf /etc/rsyslog.d/

VOLUME /var/log/docker/

EXPOSE 514

CMD cron && rsyslogd -n
