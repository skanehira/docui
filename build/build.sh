#!/bin/bash

# build docui binary
GOOS=linux GOARCH=amd64 go build ../

# build docker image
docker build -t skanehira/docui .

# remove docui binary
rm -rf ./docui

# push image to dockerr hub
docker push skanehira/docui
