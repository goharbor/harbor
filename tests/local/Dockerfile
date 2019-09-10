###
### this Dockerfile is an example representing one variant of this image;
### please see https://github.com/circleci-public/circleci-dockerfiles
### for a complete list of Dockerfiles for each tag/variant of this image
###

FROM ubuntu:xenial

# make Apt non-interactive
RUN echo 'APT::Get::Assume-Yes "true";' > /etc/apt/apt.conf.d/90harbor \
  && echo 'DPkg::Options "--force-confnew";' >> /etc/apt/apt.conf.d/90harbor

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
  && mkdir -p /usr/share/man/man1 \
  && apt-get install -y --no-install-recommends \
    apt=1.2.32 apt-transport-https=1.2.32 build-essential=12.1ubuntu2 bzip2=1.0.6-8ubuntu0.2 \
    curl=7.47.0-1ubuntu2.13 fonts-liberation=1.07.4-1 git=1:2.7.4-0ubuntu1.6 \
    gnupg=1.4.20-1ubuntu3.3 gnupg2=2.1.11-6ubuntu2.1 gzip=1.6-4ubuntu1 \
    iproute2=4.3.0-1ubuntu3.16.04.5 \
    libappindicator3-1=12.10.1+16.04.20170215-0ubuntu1 libasound2=1.1.0-0ubuntu1 \
    libatk-bridge2.0-0=2.18.1-2ubuntu1 libatk1.0-0=2.18.0-1 libatspi2.0-0=2.18.3-4ubuntu1 \
    libcairo2=1.14.6-1 libgdk-pixbuf2.0-0=2.32.2-1ubuntu1.6 libgtk-3-0=3.18.9-1ubuntu3.3 \
    libxss1=1:1.2.2-1 locales=2.23-0ubuntu11 mercurial=3.7.3-1ubuntu1.2 net-tools=1.60-26ubuntu1 \
    netcat=1.10-41 openjdk-9-jdk-headless=9~b114-0ubuntu1 openssh-client=1:7.2p2-4ubuntu2.8 \
    parallel=20161222-1~ubuntu0.16.04.1 postgresql=9.5+173ubuntu0.2 software-properties-common=0.96.20.9 \
    sudo=1.8.16-0ubuntu1.7 tar=1.28-2.1ubuntu0.1 unzip=6.0-20ubuntu1 wget=1.17.1-1ubuntu1.5 \
    xdg-utils=1.1.1-1ubuntu1.16.04.3 xvfb=2:1.18.4-0ubuntu0.8 zip=3.0-11 \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

# Set timezone to UTC by default
RUN ln -sf /usr/share/zoneinfo/Etc/UTC /etc/localtime

# Use unicode
RUN locale-gen C.UTF-8 || true
ENV LANG=C.UTF-8

# install jq
RUN JQ_URL="https://circle-downloads.s3.amazonaws.com/circleci-images/cache/linux-amd64/jq-latest" \
  && curl --silent --show-error --location --fail --retry 3 --output /usr/bin/jq "$JQ_URL" \
  && chmod +x /usr/bin/jq \
  && jq --version

# Docker & golang
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
RUN add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable"
RUN add-apt-repository ppa:longsleep/golang-backports

RUN apt-get update && apt-get install -y --no-install-recommends containerd.io=1.2.6-3 \
    docker-ce=5:19.03.2~3-0~ubuntu-xenial docker-ce-cli=5:19.03.2~3-0~ubuntu-xenial \
    golang-go=2:1.13~1longsleep1+xenial \
  && apt-get clean \
  && rm -rf /var/lib/apt/lists/*

# docker compose
RUN COMPOSE_URL="https://circle-downloads.s3.amazonaws.com/circleci-images/cache/linux-amd64/docker-compose-latest" \
  && curl --silent --show-error --location --fail --retry 3 --output /usr/bin/docker-compose "$COMPOSE_URL" \
  && chmod +x /usr/bin/docker-compose \
  && docker-compose version


USER root

RUN groupadd --gid 3434 travis \
  && useradd --uid 3434 --gid travis --groups docker --shell /bin/bash --create-home travis \
  && echo 'travis ALL=NOPASSWD: ALL' >> /etc/sudoers.d/50-travis \
  && echo 'Defaults    env_keep += "DEBIAN_FRONTEND"' >> /etc/sudoers.d/env_keep \
  && echo 'Defaults    env_keep += "REG_VERSION"' >> /etc/sudoers.d/env_keep

### Install nodejs

## Using node installation from https://raw.githubusercontent.com/nodejs/docker-node/a9c583095d4cf08bbd68f570a1f9a99780820351/10/stretch/Dockerfile

RUN groupadd --gid 1000 node \
  && useradd --uid 1000 --gid node --shell /bin/bash --create-home node

ENV NODE_VERSION 10.16.3

RUN ARCH= && dpkgArch="$(dpkg --print-architecture)" \
  && case "${dpkgArch##*-}" in \
    amd64) ARCH='x64';; \
    ppc64el) ARCH='ppc64le';; \
    s390x) ARCH='s390x';; \
    arm64) ARCH='arm64';; \
    armhf) ARCH='armv7l';; \
    i386) ARCH='x86';; \
    *) echo "unsupported architecture"; exit 1 ;; \
  esac \
  # gpg keys listed at https://github.com/nodejs/node#release-keys
  && set -ex \
  && for key in \
    94AE36675C464D64BAFA68DD7434390BDBE9B9C5 \
    FD3A5288F042B6850C66B31F09FE44734EB7990E \
    71DCFD284A79C3B38668286BC97EC7A07EDE3FC1 \
    DD8F2338BAE7501E3DD5AC78C273792F7D83545D \
    C4F0DFFF4E8C1A8236409D08E73BC641CC11F4C8 \
    B9AE9905FFD7803F25714661B63B535A4C206CA9 \
    77984A986EBC2AA786BC0F66B01FBB92821C587A \
    8FCCA13FEF1D0C2E91008E09770F7A9A5AE15600 \
    4ED778F539E3634C779C87C6D7062848A1AB005C \
    A48C2BEE680E841632CD4E44F07496B3EB3C1762 \
    B9E2F5981AA6E0CD28160D9FF13993A75599653C \
  ; do \
    gpg --batch --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys "$key" || \
    gpg --batch --keyserver hkp://ipv4.pool.sks-keyservers.net --recv-keys "$key" || \
    gpg --batch --keyserver hkp://pgp.mit.edu:80 --recv-keys "$key" ; \
  done \
  && curl -fsSLO --compressed "https://nodejs.org/dist/v$NODE_VERSION/node-v$NODE_VERSION-linux-$ARCH.tar.xz" \
  && curl -fsSLO --compressed "https://nodejs.org/dist/v$NODE_VERSION/SHASUMS256.txt.asc" \
  && gpg --batch --decrypt --output SHASUMS256.txt SHASUMS256.txt.asc \
  && grep " node-v$NODE_VERSION-linux-$ARCH.tar.xz\$" SHASUMS256.txt | sha256sum -c - \
  && tar -xJf "node-v$NODE_VERSION-linux-$ARCH.tar.xz" -C /usr/local --strip-components=1 --no-same-owner \
  && rm "node-v$NODE_VERSION-linux-$ARCH.tar.xz" SHASUMS256.txt.asc SHASUMS256.txt \
  && ln -s /usr/local/bin/node /usr/local/bin/nodejs

# install chrome

RUN curl --silent --show-error --location --fail --retry 3 --output /tmp/google-chrome-stable_current_amd64.deb https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb \
    && (dpkg -i /tmp/google-chrome-stable_current_amd64.deb || apt-get -fy --no-install-recommends install)  \
    && rm -rf /tmp/google-chrome-stable_current_amd64.deb \
    && sed -i 's|HERE/chrome"|HERE/chrome" --disable-setuid-sandbox --no-sandbox|g' \
        "/opt/google/chrome/google-chrome" \
    && ln -s "$(command -v google-chrome)" /usr/local/bin/chromium-browser \
    && chromium-browser --version

RUN CHROME_VERSION="$(google-chrome --version)" \
    && export CHROMEDRIVER_RELEASE="$(echo "$CHROME_VERSION" | sed 's/^Google Chrome //')" && export CHROMEDRIVER_RELEASE="${CHROMEDRIVER_RELEASE%%.*}" \
    && CHROMEDRIVER_VERSION="$(curl --silent --show-error --location --fail --retry 4 --retry-delay 5 "http://chromedriver.storage.googleapis.com/LATEST_RELEASE_${CHROMEDRIVER_RELEASE}")" \
    && curl --silent --show-error --location --fail --retry 4 --retry-delay 5 --output /tmp/chromedriver_linux64.zip "http://chromedriver.storage.googleapis.com/$CHROMEDRIVER_VERSION/chromedriver_linux64.zip" \
    && unzip /tmp/chromedriver_linux64.zip chromedriver -d /usr/local/bin \
    && chmod +x /usr/local/bin/chromedriver \
    && chromedriver --version

# start xvfb automatically to avoid needing to express in circle.yml
ENV DISPLAY :99
RUN printf '#!/bin/sh\nsudo Xvfb :99 -screen 0 1280x1024x24 &\nexec "$@"\n' > /tmp/entrypoint \
  && chmod +x /tmp/entrypoint \
        && mv /tmp/entrypoint /docker-entrypoint.sh

RUN mkdir -p /usr/local/lib/node_modules && chgrp -R 3434 /usr/local/ && chgrp -R 3434 /usr/lib && chmod -R g+w /usr/lib/go-1.13
RUN touch /etc/init.d/xvfb # Needed by ./tests/travis/ut_run.sh:9 :-(

COPY .env /home/travis/
COPY *.sh /home/travis/
RUN chmod +x /home/travis/*.sh

USER travis

ENV PATH $PATH:/home/travis/go/bin


VOLUME ["/h"]
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/bin/sh"]