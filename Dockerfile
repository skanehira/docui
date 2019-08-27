# build docui
FROM golang:1.12.8 AS build-docui
ENV GOOS linux
ENV GOARCH amd64
ENV GO111MODULE on
ENV CGO_ENABLED 0
COPY . ./src/github.com/skanehira/docui
WORKDIR /go/src/github.com/skanehira/docui
RUN go build

# copy artifact from the build stage
FROM busybox:1.30
ENV TERM "xterm-256color"
COPY --from=build-docui /go/src/github.com/skanehira/docui/docui /usr/local/bin/docui

ENTRYPOINT ["docui"]
