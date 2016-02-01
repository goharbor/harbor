FROM golang:1.5.1

RUN apt-get update -qqy && apt-get install -qqy libldap2-dev

COPY . /go/src/github.com/vmware/harbor
WORKDIR /go/src/github.com/vmware/harbor

ENV GOPATH /go/src/github.com/vmware/harbor/Godeps/_workspace:$GOPATH
RUN go install -v -a 

ENV MYSQL_USR root
ENV MYSQL_PWD root
ENV MYSQL_PORT_3306_TCP_ADDR localhost
ENV MYSQL_PORT_3306_TCP_PORT 3306
ENV REGISTRY_URL localhost:5000

ADD conf /go/bin/conf
ADD views /go/bin/views
ADD static /go/bin/static

RUN chmod u+x /go/bin/harbor

RUN sed -i 's/TLS_CACERT/#TLS_CAERT/g' /etc/ldap/ldap.conf
RUN sed -i '$a\TLS_REQCERT allow' /etc/ldap/ldap.conf

WORKDIR /go/bin/
ENTRYPOINT ["/go/bin/harbor"]

EXPOSE 80

