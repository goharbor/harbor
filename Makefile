.PHONY: clean build localbuild

PWD := $(shell pwd)

all: build
	docker run  --rm -it --name=harbor_container harbor_image

build:
	docker build --rm -t "harbor_image" -f Dockerfile.sry .

clean:
	@rm -rf bin
	-docker rm -f harbor_container
	-docker rmi -f harbor_image


localbuild: 
	MYSQL_USR=root MYSQL_PWD=root123 MYSQL_PORT_3306_TCP_ADDR=172.17.0.1  MYSQL_PORT_3306_TCP_PORT=3306  REDIS_HOST=10.3.10.36 REDIS_PORT=6379  REGISTRY_URL=localhost:5000 SQL_PATH=${PWD}/sql go build -v && ./harbor


