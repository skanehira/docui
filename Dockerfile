# build docui
FROM golang:1.11.5 AS build-docui
ENV GOPATH /go
ENV GOOS linux
ENV GOARCH amd64
ENV CGO_ENABLED 0
ENV GO111MODULE on
COPY . ./src/github.com/skanehira/docui
WORKDIR /go/src/github.com/skanehira/docui
RUN go build

# copy artifact from the build stage
FROM busybox:1.30
COPY --from=build-docui /go/src/github.com/skanehira/docui/docui /usr/local/bin/docui

ENTRYPOINT ["docui"]
