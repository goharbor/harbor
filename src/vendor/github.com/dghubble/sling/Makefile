.PHONY: all
all: test vet fmt

.PHONY: test
test:
	@go test . -cover

.PHONY: vet
vet:
	@go vet -all .

.PHONY: fmt
fmt:
	@test -z $$(go fmt ./...)

