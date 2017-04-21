FROM busybox:latest
MAINTAINER David Lawrence "david.lawrence@docker.com"

# the ln is for compatibility with the docker-compose.yml, making these
# images a straight swap for the those built in the compose file.
RUN mkdir -p /usr/bin /var/lib && ln -s /bin/env /usr/bin/env

COPY ./bin/notary-signer /usr/bin/notary-signer
COPY ./bin/migrate /usr/bin/migrate
COPY ./bin/ld-musl-x86_64.so.1 /lib/ld-musl-x86_64.so.1
COPY ./fixtures /var/lib/notary/fixtures
COPY ./migrations /var/lib/notary/migrations

WORKDIR /var/lib/notary
ENV SERVICE_NAME=notary_signer
ENV NOTARY_SIGNER_DEFAULT_ALIAS="timestamp_1"
ENV NOTARY_SIGNER_TIMESTAMP_1="testpassword"

ENTRYPOINT [ "/usr/bin/notary-signer" ]
CMD [ "-config=/var/lib/notary/fixtures/signer-config-local.json" ]
