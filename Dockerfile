FROM node:7.4.0

RUN mkdir -p /usr/src/app

WORKDIR /usr/src/app

COPY harbor-app /usr/src/app


RUN npm install -g bower
RUN npm install -g angular-cli
RUN npm install

EXPOSE 4200

ENTRYPOINT [ "npm", "start" ]
