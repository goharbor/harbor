FROM library/photon:1.0

MAINTAINER wangyan@vmware.com

# The original script in the docker offical registry image.
COPY entrypoint.sh /
RUN chmod u+x /entrypoint.sh

RUN mkdir -p /etc/docker/registry
COPY config.yml /etc/docker/registry/config.yml

COPY binary/registry /usr/bin
RUN chmod u+x /usr/bin/registry

VOLUME ["/var/lib/registry"]
EXPOSE 5000
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/etc/docker/registry/config.yml"]
