FROM ubuntu:xenial

RUN apt-get update && \
    apt-get install -y curl apt-transport-https

RUN curl -L https://packagecloud.io/golang-migrate/migrate/gpgkey | apt-key add - && \
    echo "deb https://packagecloud.io/golang-migrate/migrate/ubuntu/ xenial main" > /etc/apt/sources.list.d/migrate.list && \
    apt-get update && \
    apt-get install -y migrate

RUN migrate -version

