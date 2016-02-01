FROM library/ubuntu:14.04

# run logrotate hourly
RUN mv /etc/cron.daily/logrotate /etc/cron.hourly/

# logrotate configuration file for docker
ADD logrotate_docker.conf /etc/logrotate.d/

#disable imklog model
RUN sed 's/$ModLoad imklog/#$ModLoad imklog/' -i /etc/rsyslog.conf
RUN sed 's/$KLogPermitNonKernelFacility on/#$KLogPermitNonKernelFacility on/' -i /etc/rsyslog.conf

# provides TCP/UDP syslog reception
RUN sed 's/#$ModLoad imudp/$ModLoad imudp/' -i /etc/rsyslog.conf
RUN sed 's/#$UDPServerRun 514/$UDPServerRun 514/' -i /etc/rsyslog.conf
RUN sed 's/#$ModLoad imtcp/$ModLoad imtcp/' -i /etc/rsyslog.conf
RUN sed 's/#$InputTCPServerRun 514/$InputTCPServerRun 514/' -i /etc/rsyslog.conf

RUN rm /etc/rsyslog.d/*

# rsyslog configuration file for docker
ADD rsyslog_docker.conf /etc/rsyslog.d/

VOLUME /var/log/docker/

EXPOSE 514

CMD cron &&  chown -R syslog:syslog /var/log/docker/ && rsyslogd -n
