FROM ubuntu:20.04 AS runtime

ENV TZ=Asia/Shanghai \
    DEBIAN_FRONTEND=noninteractive
ENV LANG=C.UTF-8
ENV HELM_EXPERIMENTAL_OCI=1
ENV COSIGN_PASSWORD=Harbor12345
ENV COSIGN_EXPERIMENTAL=1
ENV COSIGN_OCI_EXPERIMENTAL=1
ENV NOTATION_EXPERIMENTAL=1

RUN apt-get update && apt-get install -y --no-install-recommends \
    wget \
    curl \
    gnupg2 \
    apt-transport-https \
    ca-certificates \
    lsb-release \
    libseccomp2 \
    git \
    iproute2 \
    iptables \
    build-essential \
    sed \
    libssl-dev \
    tar \
    unzip \
    gzip \
    jq \
    libnss3-tools \
    sudo

RUN wget --no-check-certificate -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add - && \
    sh -c 'echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list' && \
    apt-get update && \
    apt-get install -y --no-install-recommends google-chrome-stable

RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y nodejs && \
    npm install -g npm@latest

RUN DOCKER_VERSION=27.1.1 && \
    wget https://download.docker.com/linux/static/stable/x86_64/docker-$DOCKER_VERSION.tgz && \
    tar --strip-components=1 -xvzf docker-$DOCKER_VERSION.tgz -C /usr/bin && \
    rm docker-$DOCKER_VERSION.tgz

WORKDIR /app

RUN mkdir -p $HOME/.pki/nssdb && \
    echo Harbor12345 > password.ca && \
    certutil -d sql:$HOME/.pki/nssdb -N -f password.ca

VOLUME /var/lib/docker

CMD ["npx", "playwright", "test", "--reporter=html"]

FROM runtime AS bundled

COPY api/v2.0/swagger.yaml /api/v2.0/swagger.yaml
COPY src/portal/package.json src/portal/package-lock.json ./
COPY src/portal/scripts ./scripts
RUN npm ci && \
    npx playwright install --with-deps chromium && \
    apt-get clean all
