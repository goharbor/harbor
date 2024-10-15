ARG GOLANG
FROM ${GOLANG}

ARG SPECTRAL_VERSION
RUN case "$(dpkg --print-architecture)" in \
    amd64) ARCH="x64" ;; \
    arm64) ARCH="arm64" ;; \
    *) echo "Unsupported architecture" && exit 1 ;; \
    esac && \
    echo "Architecture: $ARCH" && \
    echo "Spectral version: $SPECTRAL_VERSION" && \
    URL="https://github.com/stoplightio/spectral/releases/download/$SPECTRAL_VERSION/spectral-linux-$ARCH" && \
    echo "URL: $URL" && \
    curl -fsSL -o /usr/bin/spectral $URL && chmod +x /usr/bin/spectral

ENTRYPOINT ["/usr/bin/spectral"]
CMD ["--version"]
