.PHONY: all
all: test vet lint fmt

.PHONY: test
test:
	@go test . -cover

.PHONY: vet
vet:
	@go vet -all .

.PHONY: lint
lint:
	@golint -set_exit_status ./...

.PHONY: fmt
fmt:
	@test -z $$(go fmt ./...)

