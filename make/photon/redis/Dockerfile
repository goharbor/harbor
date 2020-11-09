ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-redis-base:${harbor_base_image_version}

VOLUME /var/lib/redis
WORKDIR /var/lib/redis
COPY ./make/photon/redis/docker-healthcheck /usr/bin/
COPY ./make/photon/redis/docker-entrypoint.sh /usr/bin/
COPY ./make/photon/redis/redis.conf /etc/redis.conf
RUN chmod +x /usr/bin/docker-healthcheck /usr/bin/docker-entrypoint.sh \
    && chown redis:redis /etc/redis.conf
ENV REDIS_PASSWORD ""

HEALTHCHECK CMD ["docker-healthcheck"]
USER redis
ENTRYPOINT ["docker-entrypoint.sh"]
