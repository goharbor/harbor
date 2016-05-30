.PHONY: clean build localbuild

default: run

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
	GO15VENDOREXPERIMENT=1 go build -v

run: 
	MYSQL_USR=root MYSQL_PWD= MYSQL_PORT_3306_TCP_ADDR=localhost  MYSQL_PORT_3306_TCP_PORT=3306  REDIS_HOST=localhost REDIS_PORT=6379  REGISTRY_URL=http://10.3.10.36:5000 SQL_PATH=${PWD}/sql  ./harbor





