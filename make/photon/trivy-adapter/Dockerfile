ARG harbor_base_image_version
FROM goharbor/harbor-trivy-adapter-base:${harbor_base_image_version}

ARG trivy_version

COPY ./make/photon/trivy-adapter/binary/trivy /usr/local/bin/trivy
COPY ./make/photon/trivy-adapter/binary/scanner-trivy /home/scanner/bin/scanner-trivy

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl -sS 127.0.0.1:8080/probe/healthy || exit 1

ENV TRIVY_VERSION=${trivy_version}

USER scanner

ENTRYPOINT ["/home/scanner/bin/scanner-trivy"]
