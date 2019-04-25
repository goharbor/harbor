FROM golang:1.10.1

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
	python-setuptools \
	--no-install-recommends \
	&& rm -rf /var/lib/apt/lists/*

RUN useradd -ms /bin/bash notary \
	&& pip install codecov \
	&& go get github.com/golang/lint/golint github.com/fzipp/gocyclo github.com/client9/misspell/cmd/misspell github.com/gordonklaus/ineffassign github.com/HewlettPackard/gas

ENV NOTARYDIR /go/src/github.com/theupdateframework/notary

COPY . ${NOTARYDIR}
RUN chmod -R a+rw /go && chmod 0600 ${NOTARYDIR}/fixtures/database/*

WORKDIR ${NOTARYDIR}
