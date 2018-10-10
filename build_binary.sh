#!/bin/bash

# build MacOS
GOOS=darwin GOARCH=amd64 go build
zip MacOS.zip ./docui && rm -rf ./docui

# build Linux
GOOS=linux GOARCH=amd64 go build
zip Linux.zip ./docui && rm -rf ./docui

