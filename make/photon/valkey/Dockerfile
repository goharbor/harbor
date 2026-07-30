ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-valkey-base:${harbor_base_image_version}

VOLUME /var/lib/redis
WORKDIR /var/lib/redis
COPY ./make/photon/valkey/docker-healthcheck /usr/bin/
COPY ./make/photon/valkey/redis.conf /etc/redis.conf
RUN chmod +x /usr/bin/docker-healthcheck \
    && chown valkey:valkey /etc/redis.conf

HEALTHCHECK CMD ["docker-healthcheck"]
USER valkey
CMD ["valkey-server", "/etc/redis.conf"]
