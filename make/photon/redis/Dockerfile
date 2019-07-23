FROM photon:2.0

RUN tdnf install -y redis sudo 

VOLUME /var/lib/redis
WORKDIR /var/lib/redis
COPY ./make/photon/redis/docker-entrypoint.sh /usr/bin/
COPY ./make/photon/redis/redis.conf /etc/redis.conf
RUN chmod +x /usr/bin/docker-entrypoint.sh \
    && chown redis:redis /etc/redis.conf
ENTRYPOINT ["docker-entrypoint.sh"]
USER redis

EXPOSE 6379
CMD ["redis-server", "/etc/redis.conf"]
