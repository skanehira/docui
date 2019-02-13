# Go parameters
GOBUILD=go build
GOCLEAN=go clean
BINARY_NAME=docui
DOCKER_BINARY_NAME=docui-docker

export GO111MODULE=on

all: build

clean:
	$(GOCLEAN)

build: clean
	$(GOBUILD) -o $(BINARY_NAME)

# copy to $GOBIN
install: build
	cp -f $(BINARY_NAME) $(GOBIN)/

# build release binary
release: clean
	GOOS=darwin GOARCH=amd64 $(GOBUILD) && zip MacOS.zip ./docui && rm -rf ./docui
	GOOS=linux GOARCH=amd64 $(GOBUILD) && zip Linux.zip ./docui && rm -rf ./docui

docker-build:
	docker build -t skanehira/docui .

docker-push:
	docker push skanehira/docui
