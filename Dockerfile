# build docker image
FROM alpine:latest
COPY docui-docker /usr/local/bin/docui

ENTRYPOINT ["/bin/sh"]
