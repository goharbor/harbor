ARG GOLANG
FROM ${GOLANG}

ARG SWAGGER_VERSION
RUN curl -fsSL -o /usr/bin/swagger https://github.com/go-swagger/go-swagger/releases/download/$SWAGGER_VERSION/swagger_linux_$(dpkg --print-architecture) && chmod +x /usr/bin/swagger

ENTRYPOINT ["/usr/bin/swagger"]
CMD ["--help"]
