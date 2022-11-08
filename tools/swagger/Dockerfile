ARG GOLANG
FROM ${GOLANG}

ARG SWAGGER_VERSION

RUN ARCH= && dpkgArch="$(uname -m)" \
  && case "${dpkgArch}" in \
    x86_64) ARCH='x64'; curl -fsSL -o /usr/bin/swagger https://github.com/go-swagger/go-swagger/releases/download/$SWAGGER_VERSION/swagger_linux_amd64 && chmod +x /usr/bin/swagger  ;; \
    aarch64) ARCH='arm64'; curl -fsSL -o /usr/bin/swagger https://github.com/go-swagger/go-swagger/releases/download/$SWAGGER_VERSION/swagger_linux_arm64 && chmod +x /usr/bin/swagger ;; \
    ppc64el) ARCH='ppc64le' ;; \
    s390x) ARCH='s390x' ;; \
    arm64) ARCH='arm64'; curl -fsSL -o /usr/bin/swagger https://github.com/go-swagger/go-swagger/releases/download/$SWAGGER_VERSION/swagger_linux_amd64 && chmod +x /usr/bin/swagger ;; \
    armhf) ARCH='armv7l' ;; \
    i386) ARCH='x86' ;; \
    *) echo "unsupported architecture"; exit 1 ;; \
    esac

ENTRYPOINT ["/usr/bin/swagger"]
CMD ["--help"]
