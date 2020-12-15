FROM golang:1.15.6

ARG SWAGGER_VERSION
ARG TARGETARCH
RUN curl -fsSL -o /usr/bin/swagger https://github.com/go-swagger/go-swagger/releases/download/$SWAGGER_VERSION/swagger_linux_${TARGETARCH:-amd64} && chmod +x /usr/bin/swagger

ENTRYPOINT ["/usr/bin/swagger"]
CMD ["--help"]
