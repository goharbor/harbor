ARG GOLANG
FROM ${GOLANG}

ARG SWAGGER_VERSION
RUN GOBIN=/go/bin go install github.com/go-swagger/go-swagger/cmd/swagger@${SWAGGER_VERSION}

ENTRYPOINT ["swagger"]
CMD ["--help"]
