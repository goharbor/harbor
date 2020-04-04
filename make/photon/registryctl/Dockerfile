ARG harbor_base_image_version
FROM goharbor/harbor-registryctl-base:${harbor_base_image_version}

COPY ./make/photon/common/install_cert.sh /home/harbor
COPY ./make/photon/registry/binary/registry /usr/bin
COPY ./make/photon/registryctl/start.sh /home/harbor
COPY ./make/photon/registryctl/harbor_registryctl /home/harbor

RUN chown -R harbor:harbor /etc/pki/tls/certs \
    && chown harbor:harbor /home/harbor/harbor_registryctl && chmod u+x /home/harbor/harbor_registryctl \
    && chown harbor:harbor /usr/bin/registry && chmod u+x /usr/bin/registry \
    && chown harbor:harbor /home/harbor/start.sh && chmod u+x /home/harbor/start.sh \
    && chown harbor:harbor /home/harbor/install_cert.sh && chmod u+x /home/harbor/install_cert.sh


HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080/api/health || curl -k --fail -s https://127.0.0.1:8443/api/health || exit 1

VOLUME ["/var/lib/registry"]

ENTRYPOINT ["/home/harbor/start.sh"]

USER harbor
