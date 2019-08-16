FROM cicd.harbor.bitsf.xin/harbor-depend/registry:base

COPY ./make/photon/common/install_cert.sh /harbor
COPY ./make/photon/registry/entrypoint.sh /
COPY ./make/photon/registry/binary/registry /usr/bin

RUN chmod u+x /entrypoint.sh \
    && chmod u+x /usr/bin/registry

HEALTHCHECK CMD curl 127.0.0.1:5000/

VOLUME ["/var/lib/registry"]
EXPOSE 5000
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/etc/registry/config.yml"]
