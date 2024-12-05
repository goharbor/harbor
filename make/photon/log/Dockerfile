ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-log-base:${harbor_base_image_version}

COPY ./make/photon/log/rsyslog.conf /etc/rsyslog.conf

# rsyslog configuration file for docker
COPY ./make/photon/log/rsyslog_docker.conf /etc/rsyslog.d/

COPY ./make/photon/log/logrotate /etc/periodic/hourly/

COPY ./make/photon/log/start.sh /usr/local/bin/

# remove the original "logrotate" in directory "/etc/periodic/daily/"
# and copy the customized one to directory "/etc/periodic/hourly/"
# to run logrotate hourly
RUN rm /etc/periodic/daily/logrotate && \
    chmod +x /usr/local/bin/start.sh /etc/rsyslog.d/ && \
    chown -R 10000:10000 /etc/rsyslog.conf /run

HEALTHCHECK CMD netstat -ltun|grep 10514

VOLUME /var/log/docker/ /run/ /etc/logrotate.d/

CMD /usr/local/bin/start.sh
