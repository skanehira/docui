#!/bin/bash

# download docker client only
DOCKER_CLI_VERSION="18.09.1"
DOWNLOAD_URL="https://download.docker.com/linux/static/stable/x86_64/docker-$DOCKER_CLI_VERSION.tgz"

curl -L $DOWNLOAD_URL | tar -xz

# build docui binary
GOOS=linux GOARCH=amd64 go build ../

# build docker image
docker build -t skanehira/docui .

# remove docui binary
rm -rf ./docui docker

# push image to dockerr hub
docker push skanehira/docui
