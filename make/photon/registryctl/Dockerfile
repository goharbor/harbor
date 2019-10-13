FROM photon:2.0

Label maintainer="wangyan@vmware.com"

RUN tdnf install sudo -y >> /dev/null \
    && tdnf clean all \
    && groupadd -r -g 10000 harbor && useradd --no-log-init -m -g 10000 -u 10000 harbor \
    && mkdir -p /etc/registry

COPY ./make/photon/common/install_cert.sh /home/harbor
COPY ./make/photon/registry/binary/registry /usr/bin
COPY ./make/photon/registryctl/start.sh /home/harbor
COPY ./make/photon/registryctl/harbor_registryctl /home/harbor

RUN chmod -R 777 /etc/pki/tls/certs \
    && chown harbor:harbor /home/harbor/harbor_registryctl && chmod u+x /home/harbor/harbor_registryctl \
    && chown harbor:harbor /usr/bin/registry && chmod u+x /usr/bin/registry \
    && chown harbor:harbor /home/harbor/start.sh && chmod u+x /home/harbor/start.sh \
    && chown harbor:harbor /home/harbor/install_cert.sh && chmod u+x /home/harbor/install_cert.sh


HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080/api/health || exit 1

VOLUME ["/var/lib/registry"]

ENTRYPOINT ["/home/harbor/start.sh"]

USER harbor
