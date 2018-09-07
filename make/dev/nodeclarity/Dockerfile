FROM node:10.7.0

RUN mkdir -p /harbor_resources
RUN mkdir -p /harbor_src

COPY src/portal/package.json /harbor_resources
COPY make/dev/nodeclarity/entrypoint.sh /

# Install Chrome
RUN apt-get update && apt-get install -y \
	apt-transport-https \
	ca-certificates \
	curl \
	gnupg \
	hicolor-icon-theme \
	libcanberra-gtk* \
	libgl1-mesa-dri \
	libgl1-mesa-glx \
	libpango1.0-0 \
	libpulse0 \
	libv4l-0 \
	--no-install-recommends
RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
RUN dpkg -i google-chrome-stable_current_amd64.deb; apt-get -fy install
RUN rm google-chrome-stable_current_amd64.deb

# Install npm package
WORKDIR /harbor_resources

RUN npm __proxy__ install && \
    chmod u+x /entrypoint.sh
VOLUME ["/harbor_src"]
