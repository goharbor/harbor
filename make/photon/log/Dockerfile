FROM library/photon:latest

# run logrotate hourly, disable imklog model, provides TCP/UDP syslog reception
RUN tdnf install -y cronie rsyslog logrotate shadow\
    && mv /etc/cron.daily/logrotate /etc/cron.hourly/ \
	&& mkdir /etc/rsyslog.d/ \
    && mkdir /var/spool/rsyslog \
    && groupadd syslog \
    && useradd -g syslog syslog
    
ADD make/common/log/rsyslog.conf /etc/rsyslog.conf

COPY make/photon/log/logrotate.conf.photon /etc/logrotate.conf 

# logrotate configuration file for docker
ADD make/common/log/logrotate_docker.conf /etc/logrotate.d/

# rsyslog configuration file for docker
ADD make/common/log/rsyslog_docker.conf /etc/rsyslog.d/


VOLUME /var/log/docker/

EXPOSE 514

CMD crond && rsyslogd -n
