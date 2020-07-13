FROM photon:2.0

RUN tdnf install -y shadow \
    && groupadd -r -g 10000 harbor \
    && useradd --no-log-init -r -m -g 10000 -u 10000 harbor \
    && mkdir /harbor/
COPY ./make/migrations /migrations
COPY ./make/photon/standalone-db-migrator/migrate /harbor/

RUN chown -R harbor:harbor /harbor/ \
    && chown -R harbor:harbor /migrations/ \
    && chmod u+x /harbor/migrate
USER harbor

ENTRYPOINT ["/harbor/migrate"]

