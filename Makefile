# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
BINARY_NAME=docui
DOCKER_BINARY_NAME=docui-docker

all: clean build

env:
	export GO111MODULE=on

clean:
	$(GOCLEAN)

build: env
	$(GOBUILD) -o $(BINARY_NAME)

install: build
	cp -f $(BINARY_NAME) $(GOBIN)/$(BINARY_NAME)

run: build
	./$(BINARY_NAME)

# Cross compilation
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(DOCKER_BINARY_NAME)

# Docker
docker-build: build-linux
	docker build -t skanehira/docui .
	rm $(DOCKER_BINARY_NAME)

docker-push:
	docker push skanehira/docui
