FROM golang:1.20 as builder

WORKDIR /root/src

ENV CGO_ENABLED=0 \
    GO111MODULE=on \
    GOOS="linux"

COPY go.mod go.sum ./
COPY main.go ./

RUN go build -a \
    -o /go/bin/harbor-chartmuseum-migrator \
    main.go

############################

FROM alpine AS helm

ARG HELM_VERSION="v3.12.1"
RUN wget -q https://get.helm.sh/helm-${HELM_VERSION}-linux-amd64.tar.gz -O - \
    | tar -xzO linux-amd64/helm > /bin/helm && \
    chmod +x /bin/helm

############################

FROM gcr.io/distroless/static

WORKDIR /home/nonroot
USER nonroot:nonroot

COPY --from=builder /go/bin/harbor-chartmuseum-migrator /harbor-chartmuseum-migrator
COPY --from=helm /bin/helm /usr/local/bin/helm

ENTRYPOINT ["/harbor-chartmuseum-migrator"]
