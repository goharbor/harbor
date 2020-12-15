ARG harbor_base_image_version
ARG harbor_base_namespace

FROM alpine:3.12 as trivy-bin
ARG trivy_version

WORKDIR /trivy-bin

RUN set -eux; \
    \
    apk add --no-cache curl; \
    \
    case $(uname -m) in \
        x86_64) export TRIVY_DOWNLOAD_URL=https://github.com/aquasecurity/trivy/releases/download/${trivy_version}/trivy_${trivy_version//v/}_Linux-64bit.tar.gz;; \
        aarch64) export TRIVY_DOWNLOAD_URL=https://github.com/aquasecurity/trivy/releases/download/${trivy_version}/trivy_${trivy_version//v/}_Linux-ARM64.tar.gz;; \
        *) echo "unsupported architecture"; exit 1 ;; \
    esac; \
    curl --connect-timeout 30 -f -k -L ${TRIVY_DOWNLOAD_URL} | tar xvz -C /trivy-bin

FROM ${harbor_base_namespace}/harbor-trivy-adapter-base:${harbor_base_image_version}

ARG trivy_version

COPY ./make/photon/common/install_cert.sh /home/scanner
COPY ./make/photon/trivy-adapter/entrypoint.sh /home/scanner

COPY --from=trivy-bin /trivy-bin/trivy /usr/local/bin/trivy

ARG TARGETARCH
COPY ./make/photon/trivy-adapter/binary/scanner-trivy-linux-${TARGETARCH:-amd64} /home/scanner/bin/scanner-trivy

RUN chown -R scanner:scanner /etc/pki/tls/certs \
    && chown scanner:scanner /home/scanner/entrypoint.sh && chmod u+x /home/scanner/entrypoint.sh \
    && chown scanner:scanner /usr/local/bin/trivy && chmod u+x /usr/local/bin/trivy \
    && chown scanner:scanner /home/scanner/bin/scanner-trivy && chmod u+x /home/scanner/bin/scanner-trivy \
    && chown scanner:scanner /home/scanner/install_cert.sh && chmod u+x /home/scanner/install_cert.sh

HEALTHCHECK --interval=30s --timeout=10s --retries=3 CMD curl --fail -s http://127.0.0.1:8080/probe/healthy || curl -k --fail -s https://127.0.0.1:8443/probe/healthy || exit 1

ENV TRIVY_VERSION=${trivy_version}

USER scanner

ENTRYPOINT ["/home/scanner/entrypoint.sh"]
