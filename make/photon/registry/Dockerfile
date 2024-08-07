ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-registry-base:${harbor_base_image_version}

COPY --link --chown=10000:10000 --chmod=755 ./make/photon/common/install_cert.sh /home/harbor/
COPY --link --chown=10000:10000 --chmod=744 ./make/photon/registry/entrypoint.sh /home/harbor/
COPY --link --chown=10000:10000 --chmod=755 ./make/photon/registry/binary/registry /usr/bin/registry_DO_NOT_USE_GC

RUN chown -R harbor:harbor /etc/pki/tls/certs

HEALTHCHECK CMD curl --fail -s http://localhost:5000 || curl -k --fail -s https://localhost:5443 || exit 1

USER harbor

ENTRYPOINT ["/home/harbor/entrypoint.sh"]

VOLUME ["/storage"]
