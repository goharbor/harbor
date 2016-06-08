FROM library/alpine:3.3

# run logrotate hourly, disable imklog model, provides TCP/UDP syslog reception
  RUN apk add --update logrotate
  RUN mv /etc/periodic/daily/logrotate /etc/periodic/hourly/

  RUN apk add --update rsyslog \
	&& sed 's/$ModLoad imklog/#$ModLoad imklog/' -i /etc/rsyslog.conf \
	&& sed 's/$KLogPermitNonKernelFacility on/#$KLogPermitNonKernelFacility on/' -i /etc/rsyslog.conf \
	&& sed 's/#$ModLoad imudp/$ModLoad imudp/' -i /etc/rsyslog.conf \
	&& sed 's/#$UDPServerRun 514/$UDPServerRun 514/' -i /etc/rsyslog.conf \
#	&& sed 's/#$ModLoad imtcp/$ModLoad imtcp/' -i /etc/rsyslog.conf \
#	&& sed 's/#$TCPServerRun 10514/$InputTCPServerRun 10514/' -i /etc/rsyslog.conf \
	&& sed 's/$PrivDropToUser syslog/#$PrivDropToUser syslog/' -i /etc/rsyslog.conf \
	&& sed 's/$PrivDropToGroup syslog/#$PrivDropToGroup syslog/' -i /etc/rsyslog.conf \
	&& sed 's/*.emerg/#*.emerg/' -i /etc/rsyslog.conf

# logrotate configuration file for docker
ADD logrotate_docker.conf /etc/logrotate.d/

# rsyslog configuration file for docker
RUN mkdir /etc/rsyslog.d
ADD rsyslog_docker.conf /etc/rsyslog.d/

VOLUME /var/log/docker/

EXPOSE 514/udp

CMD crond && rsyslogd -n
