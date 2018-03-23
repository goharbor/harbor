FROM node:7.5.0

RUN mkdir -p /harbor_resources
RUN mkdir -p /harbor_src

COPY src/ui_ng/package.json /harbor_resources
COPY make/dev/nodeclarity/entrypoint.sh /

# Install Chrome
RUN wget -q -O - https://dl.google.com/linux/linux_signing_key.pub | apt-key add -
RUN echo "deb http://dl.google.com/linux/chrome/deb/ stable main" | tee /etc/apt/sources.list.d/google-chrome.list
RUN apt-get update && apt-get -y install google-chrome-stable

# Install npm package
WORKDIR /harbor_resources

RUN npm __proxy__ install -g @angular/cli  && \
	npm __proxy__ install && \
    chmod u+x /entrypoint.sh
VOLUME ["/harbor_src"]
