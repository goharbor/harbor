ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-registry-base:${harbor_base_image_version}

RUN --mount=type=bind,source=make/photon,target=/root \
    chown -R harbor:harbor /etc/pki/tls/certs \
    && cp /root/common/install_cert.sh /home/harbor \
    && cp /root/registry/entrypoint.sh /home/harbor \
    && cp /root/registry/binary/registry /usr/bin/registry_DO_NOT_USE_GC \
    && chown harbor:harbor /home/harbor/entrypoint.sh && chmod u+x /home/harbor/entrypoint.sh \
    && chown harbor:harbor /home/harbor/install_cert.sh && chmod u+x /home/harbor/install_cert.sh \
    && chown harbor:harbor /usr/bin/registry_DO_NOT_USE_GC && chmod u+x /usr/bin/registry_DO_NOT_USE_GC

HEALTHCHECK CMD curl --fail -s http://localhost:5000 || curl -k --fail -s https://localhost:5443 || exit 1

USER harbor

ENTRYPOINT ["/home/harbor/entrypoint.sh"]

VOLUME ["/storage"]
