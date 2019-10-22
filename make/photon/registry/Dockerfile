FROM photon:2.0

LABEL maintainer="wangyan@vmware.com"

# The original script in the docker offical registry image.
RUN tdnf install sudo -y >> /dev/null\
    && tdnf clean all \
    && mkdir -p /etc/registry \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -m -g 10000 -u 10000 harbor  

COPY ./make/photon/common/install_cert.sh /home/harbor
COPY ./make/photon/registry/entrypoint.sh /home/harbor
COPY ./make/photon/registry/binary/registry /usr/bin

RUN chown -R harbor:harbor /etc/pki/tls/certs \
    && chown harbor:harbor /home/harbor/entrypoint.sh && chmod u+x /home/harbor/entrypoint.sh \
    && chown harbor:harbor /home/harbor/install_cert.sh && chmod u+x /home/harbor/install_cert.sh \
    && chown harbor:harbor /usr/bin/registry && chmod u+x /usr/bin/registry

HEALTHCHECK CMD curl 127.0.0.1:5000/

USER harbor

ENTRYPOINT ["/home/harbor/entrypoint.sh"]

VOLUME ["/var/lib/registry"]
EXPOSE 5000
