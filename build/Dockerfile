FROM golang:alpine as gobuild

# build docui
WORKDIR /go/src
RUN apk add --no-cache \
    git \
    ca-certificates && \
    rm -rf /var/cache/apk/* && \
    go get -u github.com/skanehira/docui

# build docker image
FROM alpine:latest
COPY --from=gobuild /go/bin/docui /usr/local/bin/docui

ENTRYPOINT ["/bin/sh"]
