FROM vmware/photon:1.0

#base image for rsyslog base on photon

RUN tdnf distro-sync -y \
    && tdnf install -y cronie rsyslog shadow tar gzip \
	&& mkdir /etc/rsyslog.d/ \
    && mkdir /var/spool/rsyslog \
    && groupadd syslog \
    && useradd -g syslog syslog \
    && tdnf clean all
