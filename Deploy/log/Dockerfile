FROM library/ubuntu:14.04

# run logrotate hourly, disable imklog model, provides TCP/UDP syslog reception
RUN mv /etc/cron.daily/logrotate /etc/cron.hourly/ \
	&& sed 's/$ModLoad imklog/#$ModLoad imklog/' -i /etc/rsyslog.conf \
	&& sed 's/$KLogPermitNonKernelFacility on/#$KLogPermitNonKernelFacility on/' -i /etc/rsyslog.conf \
	&& sed 's/#$ModLoad imudp/$ModLoad imudp/' -i /etc/rsyslog.conf \
	&& sed 's/#$UDPServerRun 514/$UDPServerRun 514/' -i /etc/rsyslog.conf \
	&& sed 's/#$ModLoad imtcp/$ModLoad imtcp/' -i /etc/rsyslog.conf \
	&& sed 's/#$InputTCPServerRun 514/$InputTCPServerRun 514/' -i /etc/rsyslog.conf \
	&& sed 's/$PrivDropToUser syslog/#$PrivDropToUser syslog/' -i /etc/rsyslog.conf \
	&& sed 's/$PrivDropToGroup syslog/#$PrivDropToGroup syslog/' -i /etc/rsyslog.conf \
	&& rm /etc/rsyslog.d/*

# logrotate configuration file for docker
ADD logrotate_docker.conf /etc/logrotate.d/

# rsyslog configuration file for docker
ADD rsyslog_docker.conf /etc/rsyslog.d/

VOLUME /var/log/docker/

EXPOSE 514

CMD cron && rsyslogd -n

