# docui - TUI Client for Docker Written in Go

[![Go Report Card](https://goreportcard.com/badge/github.com/skanehira/docui?)](https://goreportcard.com/report/github.com/skanehira/docui)
[![CircleCI](https://img.shields.io/circleci/project/github/skanehira/docui.svg?style=flat-square)](https://goreportcard.com/report/github.com/skanehira/docui)
[![CircleCI](https://img.shields.io/github/release/skanehira/docui.svg?style=flat-square)](https://github.com/skanehira/docui/releases)
![GitHub All Releases](https://img.shields.io/github/downloads/skanehira/docui/total.svg?style=flat)
![GitHub commits](https://img.shields.io/github/commits-since/skanehira/docui/1.0.0.svg?style=flat-square)

## About docui
docui is TUI Client for Docker.  
docui can do as follows:

- image
    - search/pull/remove
    - save/import/load
    - inspect/filtering

- container
    - create/remove
    - start/stop
    - export/commit
    - inspect/rename/filtering
    - exec cmd

- volume
    - create/remove/prune
    - inspect/filtering

- network
    - remove
    - inspect/filtering

[![asciicast](https://asciinema.org/a/223035.svg)](https://asciinema.org/a/223035)

## Support OS
- Mac
- Linux

## Required Tools
- Go Ver.1.11.4~
- Docker Engine Ver.18.06.1~
- Git

## Installation

### From Source

If you have not installed go and set GOPATH/GOBIN,  
you must install and set env before install docui.

This project uses [Go Modules](https://github.com/golang/go/wiki/Modules) for dependencies introduced in Go 1.11.

Use go get or git clone:

```
$ go get -d github.com/skanehira/docui
$ cd $GOPATH/src/github.com/skanehira/docui
$ GO111MODULE=on go install
```

```
$ git clone https://github.com/skanehira/docui.git
$ cd docui/
$ GO111MODULE=on go install
```

Make sure your PATH includes the $GOPATH/bin directory so your commands can be easily used:

```sh
export PATH=$PATH:$GOPATH/bin
```

### Homebrew

```sh
$ brew tap skanehira/docui
$ brew install docui
```

### Snapd

```sh
$ snap install docui --classic
```

## Update

Use git pull:

```sh
$ git pull
$ GO111MODULE=on go install
```

## Log file
```sh
~/docui.log
```

## Use on Docker
```
$ docker run --rm -itv /var/run/docker.sock:/var/run/docker.sock skanehira/docui
```

## Build Docker Image
```sh
$ make docker-build
```

## Keybindings
| panel            | operation              | key                             |
| ---------------- | ---------------------- | ------------------------------- |
| all              | change panel           | <kbd>Tab</kbd>                  |
| all              | quit                   | <kbd>Ctrl</kbd> + <kbd>q</kbd>  |
| all              | quit                   | <kbd>q</kbd>                    |
| list panels      | next entry             | <kbd>j</kbd> / <kbd>↓</kbd>     |
| list  panels     | previous entry         | <kbd>k</kbd> / <kbd>↑</kbd>     |
| image list       | pull image             | <kbd>p</kbd>                    |
| image list       | search images          | <kbd>Ctrl</kbd> + <kbd>f</kbd>  |
| image list       | remove image           | <kbd>d</kbd>                    |
| image list       | create container       | <kbd>c</kbd>                    |
| image list       | inspect image          | <kbd>Enter</kbd> / <kbd>o</kbd> |
| image list       | save image             | <kbd>s</kbd>                    |
| image list       | import image           | <kbd>i</kbd>                    |
| image list       | load image             | <kbd>Ctrl</kbd> + <kbd>l</kbd>  |
| image list       | remove dangling images | <kbd>Ctrl</kbd> + <kbd>d</kbd>  |
| image list       | refresh image list     | <kbd>Ctrl</kbd> + <kbd>r</kbd>  |
| image list       | filter image           | <kbd>f</kbd>                    |
| container list   | inspect container      | <kbd>Enter</kbd> / <kbd>o</kbd> |
| container list   | remove container       | <kbd>d</kbd>                    |
| container list   | start container        | <kbd>u</kbd>                    |
| container list   | stop container         | <kbd>s</kbd>                    |
| container list   | export container       | <kbd>e</kbd>                    |
| container list   | commit container       | <kbd>c</kbd>                    |
| container list   | rename container       | <kbd>r</kbd>                    |
| container list   | refresh container list | <kbd>Ctrl</kbd> + <kbd>r</kbd>  |
| container list   | filter image           | <kbd>f</kbd>                    |
| container list   | exec container cmd     | <kbd>Ctrl</kbd> + <kbd>c</kbd>  |
| volume list      | create volume          | <kbd>c</kbd>                    |
| volume list      | remove volume          | <kbd>d</kbd>                    |
| volume list      | prune volume           | <kbd>p</kbd>                    |
| volume list      | inspect volume         | <kbd>Enter</kbd> / <kbd>o</kbd> |
| volume list      | refresh volume list    | <kbd>Ctrl</kbd> + <kbd>r</kbd>  |
| volume list      | filter image           | <kbd>f</kbd>                    |
| network list     | inspect network        | <kbd>Enter</kbd> / <kbd>o</kbd> |
| network list     | remove network         | <kbd>d</kbd>                    |
| pull image       | pull image             | <kbd>Enter</kbd>                |
| pull image       | close panel            | <kbd>Esc</kbd>                  |
| create container | next input box         | <kbd>↓</kbd>  / <kbd>Tab</kbd>  |
| create container | previous input box     | <kbd>↑</kbd>                    |
| create container | close panel            | <kbd>Esc</kbd>                  |
| detail           | cursor dwon            | <kbd>j</kbd>                    |
| detail           | cursor up              | <kbd>k</kbd>                    |
| detail           | page dwon              | <kbd>d</kbd>                    |
| detail           | page up                | <kbd>u</kbd>                    |
| search images    | search image           | <kbd>Enter</kbd>                |
| search images    | close panel            | <kbd>Esc</kbd>                  |
| images           | next image             | <kbd>j</kbd> / <kbd>↓</kbd>     |
| images           | previous image         | <kbd>k</kbd> / <kbd>↑</kbd>     |
| images           | pull image             | <kbd>Enter</kbd>                |
| images           | close panel            | <kbd>Esc</kbd>                  |
| create volume    | close panel            | <kbd>Esc</kbd>                  |
| create volume    | next input box         | <kbd>↓</kbd> / <kbd>Tab</kbd>   |
| create volume    | previous input box     | <kbd>↑</kbd>                    |


## How to use
For details of the input panel please refer to [wiki](https://github.com/skanehira/docui/blob/master/wiki.md)

## Screenshots

![](https://github.com/skanehira/docui/blob/images/images/s1.png)
![](https://github.com/skanehira/docui/blob/images/images/s2.png)
![](https://github.com/skanehira/docui/blob/images/images/s3.png)
![](https://github.com/skanehira/docui/blob/images/images/s4.png)
