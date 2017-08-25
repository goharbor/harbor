FROM library/photon:1.0

MAINTAINER wangyan@vmware.com

# The original script in the docker offical registry image.
COPY entrypoint.sh /
RUN chmod u+x /entrypoint.sh

COPY binary/registry /usr/bin
RUN chmod u+x /usr/bin/registry

VOLUME ["/var/lib/registry"]

ENV PORT 5000
EXPOSE 5000
HEALTHCHECK CMD curl http://localhost:$PORT/v2 || exit 1

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/etc/docker/registry/config.yml"]
