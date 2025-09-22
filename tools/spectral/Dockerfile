# tools/spectral/Dockerfile
ARG NODE
FROM ${NODE}

ARG SPECTRAL_VERSION  # e.g. v6.11.1 (with the leading 'v')

# Make sure we're on Debian/Ubuntu and have curl/CA certs available
RUN test -f /etc/debian_version || (echo "Debian/Ubuntu base required" && exit 1) \
 && apt-get update \
 && apt-get install -y --no-install-recommends curl ca-certificates \
 && rm -rf /var/lib/apt/lists/*

# Map Debian arch -> Spectral release arch and fetch the binary
RUN set -e; \
  case "$(dpkg --print-architecture)" in \
    amd64) ARCH="x64" ;; \
    arm64) ARCH="arm64" ;; \
    *) echo "Unsupported architecture: $(dpkg --print-architecture)"; exit 1 ;; \
  esac; \
  echo "Architecture: $ARCH"; \
  echo "Spectral version: ${SPECTRAL_VERSION}"; \
  URL="https://github.com/stoplightio/spectral/releases/download/${SPECTRAL_VERSION}/spectral-linux-${ARCH}"; \
  echo "URL: $URL"; \
  curl -fsSL -o /usr/bin/spectral "$URL"; \
  chmod +x /usr/bin/spectral

ENTRYPOINT ["/usr/bin/spectral"]
CMD ["--version"]