# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get -u
BINARY_NAME=docui
DOCKER_BINARY_NAME=docui-docker


all: clean build

module:
	GO111MODULE=on $(GOCMD) install

build: module
	$(GOBUILD) -o $(BINARY_NAME) -v

install: build
	cp $(BINARY_NAME) $(GOPATH)

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(GOPATH)/$(BINARY_NAME)

run: build
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)

# Cross compilation
build-linux:
	echo "build-linux"

# Docker
docker-build:
	GOOS=linux GOARCH=amd64 $(GOBUILD)  -o $(DOCKER_BINARY_NAME)
	docker build -t skanehira/docui .
	rm $(DOCKER_BINARY_NAME)

docker-push:
	docker push skanehira/docui