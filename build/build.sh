#!/bin/bash

# build docker image
docker build -t skanehira/docui .

# remove build image
docker rmi $(docker images --filter "dangling=true" -aq)

# push image to dockerr hub
docker push skanehira/docui
