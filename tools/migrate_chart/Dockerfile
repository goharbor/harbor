FROM python:3.8.5-slim

ENV HELM_EXPERIMENTAL_OCI=1 \
    REQUESTS_CA_BUNDLE=/etc/ssl/certs/ca-certificates.crt

ARG TARGETARCH
COPY ./migrate_chart.py ./migrate_chart.sh /

# Fetch correct Helm binary for the architecture
RUN set -e; \
    case "${TARGETARCH:-amd64}" in \
      amd64) HELM_URL="https://get.helm.sh/helm-v3.9.1-linux-amd64.tar.gz" ;; \
      arm64) HELM_URL="https://get.helm.sh/helm-v3.9.1-linux-arm64.tar.gz" ;; \
      *) echo "Unsupported arch: ${TARGETARCH}" >&2; exit 1 ;; \
    esac; \
    curl -fsSL "$HELM_URL" -o /tmp/helm.tgz; \
    tar -xzf /tmp/helm.tgz -C /tmp; \
    cp /tmp/linux-*/helm /usr/local/bin/helm; \
    chmod +x /usr/local/bin/helm; \
    rm -rf /tmp/helm*; \
    pip install --no-cache-dir click==7.1.2 requests==2.24.0 pyyaml; \
    chmod +x /migrate_chart.sh ./migrate_chart.py

ENTRYPOINT ["/migrate_chart.py"]