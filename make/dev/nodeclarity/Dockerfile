FROM node:7.5.0

RUN mkdir -p /harbor_ui

COPY src/ui_ng/package.json /harbor_ui
COPY src/ui_ng/tslint.json /harbor_ui
COPY src/ui_ng/typings.json /harbor_ui
COPY src/ui_ng/yarn.lock /harbor_ui
COPY make/dev/nodeclarity/angular-cli.json /harbor_ui
COPY make/dev/nodeclarity/entrypoint.sh /

COPY src/ui_ng/tsconfig-aot.json /harbor_ui
COPY src/ui_ng/rollup-config.js /harbor_ui

WORKDIR /harbor_ui

RUN npm __proxy__ install -g @angular/cli  && \
	npm __proxy__ install && \
    chmod u+x /entrypoint.sh
VOLUME ["/harbor_ui", "/harbor_ui/dist"]
