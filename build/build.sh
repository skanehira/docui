#!/bin/bash

# build docker image
docker build -t docui .

# remove build image
docker rmi $(docker images --filter "dangling=true" -aq)
