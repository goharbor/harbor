FROM golang:1.6.2

MAINTAINER jiangd@vmware.com

RUN apt-get update && apt-get install -y libldap2-dev \
    && rm -r /var/lib/apt/lists/*

RUN mkdir /harbor/
COPY ./make/dev/jobservice/harbor_jobservice /harbor/

RUN chmod u+x /harbor/harbor_jobservice

WORKDIR /harbor/
ENTRYPOINT ["/harbor/harbor_jobservice"]
