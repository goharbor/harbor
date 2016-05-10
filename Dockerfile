FROM testregistry.dataman.io/golang:1.5.1

MAINTAINER jiangd@vmware.com

COPY . /go/src/github.com/vmware/harbor
COPY ./vendor/golang.org /go/src/golang.org


WORKDIR /go/src/github.com/vmware/harbor

ENV GO15VENDOREXPERIMENT 1

RUN go install  -v -a

ENV MYSQL_USR=root \
    MYSQL_PWD=root \
    MYSQL_HOST=localhost \
    MYSQL_PORT=3306 \
    REDIS_HOST=192.168.1.65 \
    REDIS_PORT=6379 \
    REGISTRY_URL=registry:5000 \
    CONFIG_PATH=/etc/ui/app.conf \
    SQL_PATH=/sql

COPY views /go/bin/views
COPY static /go/bin/static
COPY CATEGORIES /go/bin/CATEGORIES

COPY ./Deploy/Omega/ui/ /etc/ui
COPY ./sql/ /sql

RUN chmod u+x /go/bin/harbor

WORKDIR /go/bin/

ENTRYPOINT ["/go/bin/harbor"]

EXPOSE 5005

