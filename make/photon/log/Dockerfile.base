FROM photon:2.0

RUN tdnf install -y cronie rsyslog logrotate shadow tar gzip sudo >> /dev/null\
    && mkdir /var/spool/rsyslog \
    && groupadd -r -g 10000 syslog && useradd --no-log-init -r -g 10000 -u 10000 syslog \
    && tdnf clean all
