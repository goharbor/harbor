lint:
	golangci-lint -v run --max-same-issues=200

build:
	go build \
        -o harbor-chartmuseum-migrator \
        main.go

docker-build:
	docker build -t goharbor/chartmuseum-migrator .
