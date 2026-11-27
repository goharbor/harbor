# syntax=docker/dockerfile:1
# check=skip=InvalidDefaultArgInFrom

ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-valkey-base:${harbor_base_image_version}

VOLUME /var/lib/redis
WORKDIR /var/lib/redis
COPY ./make/photon/valkey/docker-healthcheck /usr/bin/
COPY ./make/photon/valkey/redis.conf /etc/redis.conf
RUN chmod +x /usr/bin/docker-healthcheck \
    && chown valkey:valkey /etc/redis.conf

USER valkey

HEALTHCHECK  --interval=30s --timeout=5s --start-period=10s --retries=3 \ 
    CMD ["docker-healthcheck"]

CMD ["valkey-server", "/etc/redis.conf"]
