FROM node:7.5.0

RUN mkdir -p /harbor_resources
RUN mkdir -p /harbor_src

COPY make/dev/npm-installer/entrypoint.sh /

WORKDIR /harbor_resources

RUN chmod a+x /entrypoint.sh
VOLUME ["/harbor_src"]
