FROM ubuntu:18.04
ENV LANG C.UTF-8
# V 1.9

RUN apt-get update && apt-get install -y --no-install-recommends wget curl gnupg2
RUN apt-get install libseccomp2
RUN wget  --no-check-certificate -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
RUN sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list'

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

RUN apt-get update && apt-get install -y software-properties-common && \
    add-apt-repository -y ppa:longsleep/golang-backports
RUN apt-get update && \
    apt-get install -y  golang-go

RUN apt-get update -y ; apt-get install -y  python3.6
RUN rm /usr/bin/python ; ln -s /usr/bin/python3.6 /usr/bin/python ; apt-get install -y python3-pip
RUN python -m pip install --upgrade pip

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

#RUN pip install pyasn1 google-apitools==0.5.15 gsutil robotframework==2.9.2 robotframework-sshlibrary robotframework-httplibrary requests dbbot robotframework-selenium2library robotframework-pabot robotframework-JSONLibrary --upgrade

RUN pip install pyasn1 robotframework==3.0.1 robotframework-sshlibrary robotframework-httplibrary requests dbbot robotframework-selenium2library robotframework-pabot robotframework-JSONLibrary --upgrade

# Install docker, docker compose
ENV DOCKER_VERSION 19.03.12
RUN wget https://download.docker.com/linux/static/stable/x86_64/docker-$DOCKER_VERSION.tgz && \
    tar --strip-components=1 -xvzf docker-$DOCKER_VERSION.tgz -C /usr/bin &&  \
    curl -L "https://github.com/docker/compose/releases/download/1.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose && \
    chmod +x /usr/local/bin/docker-compose

RUN wget https://github.com/drone/drone-cli/releases/download/v0.8.3/drone_linux_amd64.tar.gz && tar zxf drone_linux_amd64.tar.gz && \
    install -t /usr/local/bin drone && mv drone drone_src

RUN set -x \
        && groupadd --system dockremap \
        && adduser --system --ingroup dockremap dockremap \
        && echo 'dockremap:165536:65536' >> /etc/subuid \
        && echo 'dockremap:165536:65536' >> /etc/subgid

RUN curl -sSL https://github.com/vmware/govmomi/releases/download/v0.16.0/govc_linux_amd64.gz | gzip -d > /usr/local/bin/govc && \
    chmod +x /usr/local/bin/govc

RUN wget https://get.helm.sh/helm-v2.16.12-linux-amd64.tar.gz && tar zxvf helm-v2.16.12-linux-amd64.tar.gz && \
    cp linux-amd64/helm /usr/local/bin/helm && \
    cp linux-amd64/helm /usr/local/bin/helm2 && \
    helm init --stable-repo-url https://charts.helm.sh/stable --client-only && \
    helm plugin install https://github.com/chartmuseum/helm-push

RUN wget https://get.helm.sh/helm-v3.3.3-linux-amd64.tar.gz && tar zxvf helm-v3.3.3-linux-amd64.tar.gz && \
    mv linux-amd64/helm /usr/local/bin/helm3 && \
    helm3 plugin install https://github.com/chartmuseum/helm-push

RUN curl -LO https://github.com/deislabs/oras/releases/download/v0.9.0/oras_0.9.0_linux_amd64.tar.gz && \
    mkdir -p oras-install/  && \
    tar -zxf oras_0.9.0_*.tar.gz -C oras-install/  && \
    mv oras-install/oras /usr/local/bin/

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


RUN apt-get install -y sudo uuid-dev


ENV DIND_COMMIT 3b5fac462d21ca164b3778647420016315289034

RUN wget "https://raw.githubusercontent.com/docker/docker/${DIND_COMMIT}/hack/dind" -O /usr/local/bin/dind \
    && chmod +x /usr/local/bin/dind

COPY busybox.tar /

# This container needs to be run in privileged mode(run with --privileged option) to make it work
COPY dockerd-entrypoint.sh /usr/local/bin/dockerd-entrypoint.sh

RUN chmod +x /usr/local/bin/dockerd-entrypoint.sh

VOLUME /var/lib/docker
