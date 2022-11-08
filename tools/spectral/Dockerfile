ARG GOLANG
FROM ${GOLANG}

ARG SPECTRAL_VERSION

RUN ARCH= && dpkgArch="$(uname -m)" \
  && case "${dpkgArch}" in \
    x86_64) ARCH='x64'; curl -fsSL -o /usr/bin/spectral https://github.com/stoplightio/spectral/releases/download/$SPECTRAL_VERSION/spectral-linux && chmod +x /usr/bin/spectral ;; \
    aarch64) ARCH='arm64'; curl -fsSL -o /usr/bin/spectral https://harbor-dev-spec.oss-cn-beijing.aliyuncs.com/spectral && chmod +x /usr/bin/spectral ;; \
    ppc64el) ARCH='ppc64le'; ;; \
    s390x) ARCH='s390x'; ;; \
    armhf) ARCH='armv7l'; ;; \
    i386) ARCH='x86'; ;; \
    *) echo "unsupported architecture"; exit 1 ;; \
    esac

ENTRYPOINT ["/usr/bin/spectral"]
CMD ["--version"]
