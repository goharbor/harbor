FROM vmware/photon:1.0

#base image for rsyslog base on photon

RUN tdnf distro-sync -y || echo \
    && tdnf install -y cronie rsyslog shadow tar gzip \
	&& mkdir /etc/rsyslog.d/ \
    && mkdir /var/spool/rsyslog \
    && groupadd -r -g 10000 syslog && useradd --no-log-init -r -g 10000 -u 10000 syslog \
    && tdnf clean all
