ARG harbor_base_image_version
ARG harbor_base_namespace
FROM ${harbor_base_namespace}/harbor-trivy-adapter-base:${harbor_base_image_version}

ARG trivy_version

COPY --link --chown=10000:10000 --chmod=755 ./make/photon/common/install_cert.sh /home/scanner/
COPY --link --chown=10000:10000 --chmod=744 ./make/photon/trivy-adapter/entrypoint.sh /home/scanner/
COPY --link --chown=10000:10000 --chmod=755 ./make/photon/trivy-adapter/binary/trivy /usr/local/bin/
COPY --link --chown=10000:10000 --chmod=755 ./make/photon/trivy-adapter/binary/scanner-trivy /home/scanner/bin/


RUN chown -R scanner:scanner /etc/pki/tls/certs

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl --fail -s http://localhost:8080/probe/healthy || curl -k --fail -s https://localhost:8443/probe/healthy || exit 1

ENV TRIVY_VERSION=${trivy_version}

USER scanner

ENTRYPOINT ["/home/scanner/entrypoint.sh"]
