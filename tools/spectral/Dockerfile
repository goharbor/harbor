ARG GOLANG
FROM ${GOLANG}

ARG SPECTRAL_VERSION
RUN case "$(dpkg --print-architecture)" in \
    amd64) ARCH="x86" ;; \
    arm64) ARCH="arm64" ;; \
    *) echo "Unsupported architecture" && exit 1 ;; \
    esac && \
    curl -fsSL -o /usr/bin/spectral https://github.com/stoplightio/spectral/releases/download/$SPECTRAL_VERSION/spectral-linux-$ARCH && chmod +x /usr/bin/spectral

ENTRYPOINT ["/usr/bin/spectral"]
CMD ["--version"]
