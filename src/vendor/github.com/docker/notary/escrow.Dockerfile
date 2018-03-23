FROM golang:1.7.3-alpine
MAINTAINER David Lawrence "david.lawrence@docker.com"

ENV NOTARYPKG github.com/docker/notary

# Copy the local repo to the expected go path
COPY . /go/src/${NOTARYPKG}

WORKDIR /go/src/${NOTARYPKG}

EXPOSE 4450

# Install escrow
RUN go install ${NOTARYPKG}/cmd/escrow

ENTRYPOINT [ "escrow" ]
CMD [ "-config=cmd/escrow/config.toml" ]
