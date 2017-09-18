FROM node:7.5.0

RUN mkdir -p /harbor_resources
RUN mkdir -p /harbor_src

COPY src/ui_ng/package.json /harbor_resources
COPY make/dev/nodeclarity/entrypoint.sh /

WORKDIR /harbor_resources

RUN npm __proxy__ install -g @angular/cli  && \
	npm __proxy__ install && \
    chmod u+x /entrypoint.sh
VOLUME ["/harbor_src"]
