FROM library/ubuntu:14.04

MAINTAINER jiangd@vmware.com

RUN mkdir /harbor/
COPY ./make/dev/jobservice/harbor_jobservice /harbor/

RUN chmod u+x /harbor/harbor_jobservice

WORKDIR /harbor/
ENTRYPOINT ["/harbor/harbor_jobservice"]
