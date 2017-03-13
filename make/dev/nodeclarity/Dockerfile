FROM node:7.5.0

COPY angular-cli.json /
COPY index.html /
COPY entrypoint.sh /

RUN npm install -g @angular/cli && \
    chmod u+x entrypoint.sh

VOLUME ["/clarity-seed", "/clarity-seed/dist"]

ENTRYPOINT ["/entrypoint.sh"]
