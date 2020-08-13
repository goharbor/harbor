FROM ubuntu:18.04
ENV LANG C.UTF-8


RUN apt-get update && apt-get install -y --no-install-recommends \
    jq \
    bc \
    time \
    gcc \
    git \
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
    wget curl \
	parted \
    #ip tool
    #ethtool \
    iproute2 \
    #bridge-utils \
    #iputils-ping \
    # Add docker in docker support
    btrfs-tools \
    e2fsprogs \
    iptables \
    xfsprogs \
    dnsutils \
    netcat \
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

RUN apt-get update && \
    apt-get install golang -y

RUN apt-get update -y && \
	apt-get install -y zbar-tools libzbar-dev python-zbar
RUN dpkg -L libzbar-dev; ls -l /usr/include/zbar.h
RUN apt-get update -y
RUN apt-get install -y python3.6
RUN rm /usr/bin/python
RUN ln -s /usr/bin/python3.6 /usr/bin/python
RUN apt-get install -y python3-pip



RUN apt-get update && apt install libnss3-tools && \
    mkdir -p $HOME/.pki/nssdb && \
    echo Harbor12345 > password.ca && \
    certutil -d sql:$HOME/.pki/nssdb -N -f password.ca

RUN pip3 install pyasn1 google-apitools==0.5.31 gsutil robotframework==3.2.1 robotframework-sshlibrary robotframework-httplibrary requests dbbot robotframework-seleniumlibrary==4.3.0 robotframework-pabot robotframework-JSONLibrary --upgrade

# Install docker, docker compose
RUN wget https://download.docker.com/linux/static/stable/x86_64/docker-18.06.3-ce.tgz && \
    tar --strip-components=1 -xvzf docker-18.06.3-ce.tgz -C /usr/bin &&  \
    curl -L "https://github.com/docker/compose/releases/download/1.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose && \
    chmod +x /usr/local/bin/docker-compose

RUN wget https://github.com/drone/drone-cli/releases/download/v0.8.3/drone_linux_amd64.tar.gz && tar zxf drone_linux_amd64.tar.gz && \
    install -t /usr/local/bin drone && mv drone drone_src

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
