FROM golang:1.11.2

RUN wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
RUN sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list'

#RUN wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb && apt install ./google-chrome-stable_current_amd64.deb

RUN apt-get update && apt-get install -y --no-install-recommends \
    jq \
    bc \
    time \
    gcc \
    python-dev \
    libffi-dev \
    libssl-dev \
    sshpass \
    ant \
	ant-optional \
	xvfb \
    libxi6 \
    libgconf-2-4 \
    openjdk-8-jdk \
    rpcbind \
    nfs-common \
    unzip \
    zip \
    bzip2 \
	parted \
    # Add docker in docker support
    btrfs-tools \
    e2fsprogs \
    iptables \
    xfsprogs \
    dnsutils \
    netcat \
    # Add headless chrome support
    google-chrome-stable \
    # Speed up ISO builds with already installed reqs
    yum \
    yum-utils \
    cpio \
    rpm \
    ca-certificates \
    xz-utils \
    xorriso \
    sendmail && \
	# Cleanup
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

RUN wget -N http://chromedriver.storage.googleapis.com/2.40/chromedriver_linux64.zip && \
    unzip chromedriver_linux64.zip && \
    chmod +x chromedriver && \
    mv -f chromedriver /usr/local/share/chromedriver && \
    ln -s /usr/local/share/chromedriver /usr/local/bin/chromedriver && \
    ln -s /usr/local/share/chromedriver /usr/bin/chromedriver

RUN apt-get update && apt install libnss3-tools && \
    mkdir -p $HOME/.pki/nssdb && \
    echo Harbor12345 > password.ca && \
    certutil -d sql:$HOME/.pki/nssdb -N -f password.ca

RUN wget https://bootstrap.pypa.io/get-pip.py && \
    python ./get-pip.py  && \
    pip install pyasn1 google-apitools==0.5.15 gsutil robotframework==3.0.4 robotframework-sshlibrary robotframework-httplibrary requests dbbot robotframework-selenium2library robotframework-pabot robotframework-JSONLibrary --upgrade

# Install docker, docker compose
RUN wget https://download.docker.com/linux/static/stable/x86_64/docker-18.06.3-ce.tgz && \
    tar --strip-components=1 -xvzf docker-18.06.3-ce.tgz -C /usr/bin &&  \
    curl -L "https://github.com/docker/compose/releases/download/1.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose && \
    chmod +x /usr/local/bin/docker-compose

RUN wget https://github.com/drone/drone-cli/releases/download/v0.8.3/drone_linux_amd64.tar.gz && tar zxf drone_linux_amd64.tar.gz && \
    install -t /usr/local/bin drone

RUN curl -sSL https://github.com/vmware/govmomi/releases/download/v0.16.0/govc_linux_amd64.gz | gzip -d > /usr/local/bin/govc && \
    chmod +x /usr/local/bin/govc

RUN wget https://get.helm.sh/helm-v2.14.1-linux-386.tar.gz && tar zxvf helm-v2.14.1-linux-386.tar.gz && \
    mv linux-386/helm /usr/local/bin/helm && \
	helm init --client-only && \
	helm plugin install https://github.com/chartmuseum/helm-push

RUN wget https://get.helm.sh/helm-v3.0.0-linux-386.tar.gz && tar zxvf helm-v3.0.0-linux-386.tar.gz && \
    mv linux-386/helm /usr/local/bin/helm3 && \
	helm3 plugin install https://github.com/chartmuseum/helm-push

RUN wget https://github.com/theupdateframework/notary/releases/download/v0.6.1/notary-Linux-amd64 && \
    chmod +x notary-Linux-amd64 && \
    mv notary-Linux-amd64 /usr/local/bin/notary

# Add docker in docker support
# version: docker:1.13-dind
# reference: https://github.com/docker-library/docker/blob/b202ec7e529f5426e2ad7e8c0a8b82cacd406573/1.13/dind/Dockerfile
#
# https://github.com/docker/docker/blob/master/project/PACKAGERS.md#runtime-dependencies

# set up subuid/subgid so that "--userns-remap=default" works out-of-the-box
RUN set -x \
        && groupadd --system dockremap \
        && adduser --system --ingroup dockremap dockremap \
        && echo 'dockremap:165536:65536' >> /etc/subuid \
        && echo 'dockremap:165536:65536' >> /etc/subgid

ENV DIND_COMMIT 3b5fac462d21ca164b3778647420016315289034

RUN wget "https://raw.githubusercontent.com/docker/docker/${DIND_COMMIT}/hack/dind" -O /usr/local/bin/dind \
        && chmod +x /usr/local/bin/dind

# This container needs to be run in privileged mode(run with --privileged option) to make it work
COPY dockerd-entrypoint.sh /usr/local/bin/dockerd-entrypoint.sh
RUN chmod +x /usr/local/bin/dockerd-entrypoint.sh

VOLUME /var/lib/docker
