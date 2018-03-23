FROM golang:1.7.3

RUN apt-get update && apt-get install -y \
	curl \
	clang \
	libltdl-dev \
	libsqlite3-dev \
	patch \
	tar \
	xz-utils \
	python \
	python-pip \
	--no-install-recommends \
	&& rm -rf /var/lib/apt/lists/*

RUN useradd -ms /bin/bash notary \
	&& pip install codecov \
	&& go get github.com/golang/lint/golint github.com/fzipp/gocyclo github.com/client9/misspell/cmd/misspell github.com/gordonklaus/ineffassign github.com/HewlettPackard/gas

ENV NOTARYDIR /go/src/github.com/docker/notary

COPY . ${NOTARYDIR}
RUN chmod -R a+rw /go

WORKDIR ${NOTARYDIR}
