FROM golang:1.10.1-alpine

ENV NOTARYPKG github.com/theupdateframework/notary

# Copy the local repo to the expected go path
COPY . /go/src/${NOTARYPKG}

WORKDIR /go/src/${NOTARYPKG}

EXPOSE 4450

# Install escrow
RUN go install ${NOTARYPKG}/cmd/escrow

ENTRYPOINT [ "escrow" ]
CMD [ "-config=cmd/escrow/config.toml" ]
