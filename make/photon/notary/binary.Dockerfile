FROM golang:1.7.3

ENV NOTARY_DIR /go/src/github.com/docker/notary
ENV NOTARYPKG github.com/docker/notary

COPY . /go/src/${NOTARYPKG}
WORKDIR /go/src/${NOTARYPKG}

RUN go build -tags pkcs11 \
    -ldflags "-w -X ${NOTARYPKG}/version.GitCommit=`git rev-parse --short HEAD` -X ${NOTARYPKG}/version.NotaryVersion=`cat NOTARY_VERSION`" $NOTARYPKG/cmd/notary-server 

RUN go build -tags pkcs11 \
    -ldflags "-w -X ${NOTARYPKG}/version.GitCommit=`git rev-parse --short HEAD` -X ${NOTARYPKG}/version.NotaryVersion=`cat NOTARY_VERSION`" $NOTARYPKG/cmd/notary-signer
