FROM vmware/photon:1.0

MAINTAINER wangyan@vmware.com

# The original script in the docker offical registry image.
RUN tdnf distro-sync -y \
    && tdnf erase vim -y \
    && tdnf install sudo -y >> /dev/null\
    && tdnf clean all \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -r -g 10000 -u 10000 harbor

COPY entrypoint.sh /
RUN chmod u+x /entrypoint.sh

RUN mkdir -p /etc/registry

COPY binary/registry /usr/bin
RUN chmod u+x /usr/bin/registry

HEALTHCHECK CMD curl 127.0.0.1:5000/

VOLUME ["/var/lib/registry"]
EXPOSE 5000
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/etc/registry/config.yml"]
