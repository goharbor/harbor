FROM vmware/photon:1.0

RUN tdnf distro-sync -y \
    && tdnf install -y redis sudo \
    && mkdir /data \
    && chown redis:redis /data

VOLUME /data
WORKDIR /data
COPY docker-entrypoint.sh /usr/bin/
COPY redis.conf /etc/redis.conf
RUN chmod +x /usr/bin/docker-entrypoint.sh \
    && chown redis:redis /etc/redis.conf
ENTRYPOINT ["docker-entrypoint.sh"]

EXPOSE 6379
CMD ["redis-server", "/etc/redis.conf"]
