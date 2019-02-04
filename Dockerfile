# build docker image
FROM alpine:latest
COPY docui-docker /usr/local/bin/docui

ENTRYPOINT ["/usr/local/bin/docui"]
