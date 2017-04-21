FROM node:7.5.0

RUN mkdir -p /clarity-seed

COPY src/ui_ng/package.json /clarity-seed
COPY src/ui_ng/tslint.json /clarity-seed
COPY src/ui_ng/typings.json /clarity-seed
COPY src/ui_ng/yarn.lock /clarity-seed
COPY make/dev/nodeclarity/angular-cli.json /clarity-seed
COPY make/dev/nodeclarity/entrypoint.sh /

WORKDIR /clarity-seed

RUN npm install -g @angular/cli  && \
	npm install && \
    chmod u+x /entrypoint.sh

VOLUME ["/clarity-seed", "/clarity-seed/dist"]