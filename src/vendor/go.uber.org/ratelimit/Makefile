# Directory to put `go install`ed binaries in.
export GOBIN ?= $(shell pwd)/bin

GO_FILES := $(shell \
	find . '(' -path '*/.*' -o -path './vendor' ')' -prune \
	-o -name '*.go' -print | cut -b3-)

.PHONY: bench
bench: bin/benchstat bin/benchart
	go test -timeout 3h -count=5 -run=xxx -bench=BenchmarkRateLimiter ./... | tee stat.txt
	@$(GOBIN)/benchstat stat.txt
	@$(GOBIN)/benchstat -csv stat.txt > stat.csv
	@$(GOBIN)/benchart 'RateLimiter;xAxisType=log' stat.csv stat.html
	@open stat.html

bin/benchstat: tools/go.mod
	@cd tools && go install golang.org/x/perf/cmd/benchstat

bin/benchart: tools/go.mod
	@cd tools && go install github.com/storozhukBM/benchart

bin/golint: tools/go.mod
	@cd tools && go install golang.org/x/lint/golint

bin/staticcheck: tools/go.mod
	@cd tools && go install honnef.co/go/tools/cmd/staticcheck

.PHONY: build
build:
	go build ./...

.PHONY: cover
cover:
	go test -coverprofile=cover.out -coverpkg=./... -v ./...
	go tool cover -html=cover.out -o cover.html

.PHONY: gofmt
gofmt:
	$(eval FMT_LOG := $(shell mktemp -t gofmt.XXXXX))
	@gofmt -e -s -l $(GO_FILES) > $(FMT_LOG) || true
	@[ ! -s "$(FMT_LOG)" ] || (echo "gofmt failed:" | cat - $(FMT_LOG) && false)

.PHONY: golint
golint: bin/golint
	@$(GOBIN)/golint -set_exit_status ./...

.PHONY: lint
lint: gofmt golint staticcheck

.PHONY: staticcheck
staticcheck: bin/staticcheck
	@$(GOBIN)/staticcheck ./...

.PHONY: test
test:
	go test -race ./...
