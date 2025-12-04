# tools/spectral/Dockerfile

ARG NODE
FROM ${NODE}

ARG SPECTRAL_VERSION

# Map arch → Spectral binary name and fetch
RUN case "$(dpkg --print-architecture)" in \
      amd64) ARCH="x64" ;; \
      arm64) ARCH="arm64" ;; \
      *) echo "Unsupported architecture: $(dpkg --print-architecture)" && exit 1 ;; \
    esac && \
    echo "Architecture: $ARCH" && \
    echo "Spectral version: $SPECTRAL_VERSION" && \
    URL="https://github.com/stoplightio/spectral/releases/download/$SPECTRAL_VERSION/spectral-linux-$ARCH" && \
    echo "URL: $URL" && \
    curl -fsSL -o /usr/bin/spectral "$URL" && chmod +x /usr/bin/spectral

ENTRYPOINT ["/usr/bin/spectral"]
CMD ["--version"]