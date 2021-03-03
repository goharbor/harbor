FROM ubuntu:18.04
ENV LANG C.UTF-8
# V 2.0
# V 2.0.1: upgrade docker to version 19.03.12
# V 2.5 Add support for e2e py-test (especially containerd)
# V 2.6 docker 19.03.12
# V 2.6.1 upgrade helm2, helm3 and docker 20.10.1

RUN apt-get update && apt-get install -y --no-install-recommends wget curl gnupg2
RUN apt-get install libseccomp2
RUN wget  --no-check-certificate -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
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

RUN sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list && apt-get clean
RUN apt-get update -y && \
	apt-get install -y zbar-tools libzbar-dev python-zbar
RUN dpkg -L libzbar-dev; ls -l /usr/include/zbar.h
RUN apt-get update -y
RUN apt-get install -y python3.5
RUN rm /usr/bin/python
RUN ln -s /usr/bin/python3.5 /usr/bin/python
RUN apt-get install -y python3-pip

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

RUN pip3 install pyasn1 google-apitools==0.5.31 gsutil robotframework==3.2.1 robotframework-sshlibrary robotframework-httplibrary requests dbbot robotframework-seleniumlibrary==4.3.0 robotframework-pabot robotframework-JSONLibrary --upgrade

ENV CRI_CONTAINERD_VERSION 1.3.4
RUN wget https://storage.googleapis.com/cri-containerd-release/cri-containerd-${CRI_CONTAINERD_VERSION}.linux-amd64.tar.gz && \
    tar --no-overwrite-dir -C / -xzf cri-containerd-${CRI_CONTAINERD_VERSION}.linux-amd64.tar.gz

# Install docker, docker compose
RUN wget https://download.docker.com/linux/static/stable/x86_64/docker-19.03.12.tgz && \
    tar --strip-components=1 -xvzf docker-19.03.12.tgz -C /usr/bin &&  \
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

RUN wget http://nchc.dl.sourceforge.net/sourceforge/tcl/tcl8.4.11-src.tar.gz && \
    tar xfvz tcl8.4.11-src.tar.gz && \
    cd tcl8.4.11/unix && \
    sed -i "s/relid'/relid/" configure && \
    ./configure --prefix=/ixdba.net && \
    make && \
    make install && \
    cp -r * ../tools && \
    cd .. && \
    cp -r tools / && \
    cd /  && \
    wget http://sourceforge.net/projects/expect/files/Expect/5.45/expect5.45.tar.gz && \
    tar xzvf expect5.45.tar.gz && \
    cd expect5.45 && \
    ./configure --prefix=/tools --with-tcl=/tools --with-x=no && \
    make && \
    make install && \
    cd /ixdba.net/bin && \
    mv  expect /usr/local/bin/expect

RUN CNAB_PATH=$(go env GOPATH)/src/github.com/docker && mkdir -p $CNAB_PATH && cd $CNAB_PATH && git clone https://github.com/cnabio/cnab-to-oci.git && \
    cd cnab-to-oci && git checkout v0.3.0-beta4 && \
    go list && \
    make build && \
    mv bin/cnab-to-oci /usr/local/bin

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
