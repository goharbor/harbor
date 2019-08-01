FROM photon:2.0

RUN tdnf install -y redis sudo 

VOLUME /var/lib/redis
WORKDIR /var/lib/redis
COPY ./make/photon/redis/docker-healthcheck /usr/bin/
COPY ./make/photon/redis/redis.conf /etc/redis.conf
RUN chmod +x /usr/bin/docker-healthcheck \
    && chown redis:redis /etc/redis.conf

HEALTHCHECK CMD ["docker-healthcheck"]
USER redis
EXPOSE 6379
CMD ["redis-server", "/etc/redis.conf"]
