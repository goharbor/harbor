.PHONY: clean build

all: build
	docker run  --rm -it --name=harbor_container harbor_image

build:
	docker build --rm -t "harbor_image" -f Dockerfile.sry .

clean:
	@rm -rf bin
	-docker rm -f harbor_container
	-docker rmi -f harbor_image

