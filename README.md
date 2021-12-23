# docui - TUI Client for Docker Written in Go

[![Go Report Card](https://goreportcard.com/badge/github.com/skanehira/docui?)](https://goreportcard.com/report/github.com/skanehira/docui)
[![CircleCI](https://img.shields.io/circleci/project/github/skanehira/docui.svg?style=flat-square)](https://goreportcard.com/report/github.com/skanehira/docui)
[![CircleCI](https://img.shields.io/github/release/skanehira/docui.svg?style=flat-square)](https://github.com/skanehira/docui/releases)
![GitHub All Releases](https://img.shields.io/github/downloads/skanehira/docui/total.svg?style=flat)
![GitHub commits](https://img.shields.io/github/commits-since/skanehira/docui/1.0.0.svg?style=flat-square)

# This repository is no longer maintenance. Please use [lazydocker](https://github.com/jesseduffield/lazydocker) instead.

## About docui
![demo](https://github.com/skanehira/docui/blob/images/images/docui.v2-demo.gif?raw=true)

docui is a TUI Client for Docker.
It can do the following:

- image
    - search/pull/remove
    - save/import/load
    - inspect/filtering

- container
    - create/remove
    - start/stop/kill
    - export/commit
    - inspect/rename/filtering
    - exec cmd

- volume
    - create/remove
    - inspect/filtering

- network
    - remove
    - inspect/filtering

## Supported OSes
- Mac
- Linux

## Required Tools
- Go Ver.1.11.4~
- Docker Engine Ver.18.06.1~
- Git

## Installation
### Environment variables
The following environment variables must be set.

```
export LC_CTYPE=en_US.UTF-8
export TERM=xterm-256color
```

### From Source

If you have not installed go and set GOPATH/GOBIN,
you must install and set env before installing docui.

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
$ brew install docui
```

### Nix

docui is available on nixpkgs unstable channel:

```sh
$ nix-env -i docui
```

## Update

Use git pull:

```sh
$ git pull
$ GO111MODULE=on go install
```

## Log file

Check [wiki](https://github.com/skanehira/docui/blob/master/wiki.md).

## Use on Docker
```
$ docker run --rm -itv /var/run/docker.sock:/var/run/docker.sock skanehira/docui
```

## Build Docker Image
```sh
$ make docker-build
```

## Keybindings
| panel            | operation              | key                                                |
|------------------|------------------------|----------------------------------------------------|
| all              | change panel           | <kbd>Tab</kbd> / <kbd>Shift</kbd> + <kbd>Tab</kbd> |
| all              | quit                   | <kbd>q</kbd>                                       |
| list panels      | next entry             | <kbd>j</kbd> / <kbd>↓</kbd>                        |
| list panels      | previous entry         | <kbd>k</kbd> / <kbd>↑</kbd>                        |
| list panels      | next page              | <kbd>Ctrl</kbd> / <kbd>f</kbd>                     |
| list panels      | previous page          | <kbd>Ctrl</kbd> / <kbd>b</kbd>                     |
| list panels      | scroll to top          | <kbd>g</kbd>                                       |
| list panels      | scroll to bottom       | <kbd>G</kbd>                                       |
| image list       | pull image             | <kbd>p</kbd>                                       |
| image list       | search images          | <kbd>f</kbd>                                       |
| image list       | remove image           | <kbd>d</kbd>                                       |
| image list       | create container       | <kbd>c</kbd>                                       |
| image list       | inspect image          | <kbd>Enter</kbd>                                   |
| image list       | save image             | <kbd>s</kbd>                                       |
| image list       | import image           | <kbd>i</kbd>                                       |
| image list       | load image             | <kbd>Ctrl</kbd> + <kbd>l</kbd>                     |
| image list       | refresh image list     | <kbd>Ctrl</kbd> + <kbd>r</kbd>                     |
| image list       | filter image           | <kbd>/</kbd>                                       |
| container list   | inspect container      | <kbd>Enter</kbd>                                   |
| container list   | remove container       | <kbd>d</kbd>                                       |
| container list   | start container        | <kbd>u</kbd>                                       |
| container list   | stop container         | <kbd>s</kbd>                                       |
| container list   | kill container         | <kbd>Ctrl</kbd> + <kbd>k</kbd>                     |
| container list   | export container       | <kbd>e</kbd>                                       |
| container list   | commit container       | <kbd>c</kbd>                                       |
| container list   | rename container       | <kbd>r</kbd>                                       |
| container list   | refresh container list | <kbd>Ctrl</kbd> + <kbd>r</kbd>                     |
| container list   | filter image           | <kbd>/</kbd>                                       |
| container list   | exec container cmd     | <kbd>Ctrl</kbd> + <kbd>e</kbd>                     |
| container logs   | show container logs    | <kbd>Ctrl</kbd> + <kbd>l</kbd>                     |
| volume list      | create volume          | <kbd>c</kbd>                                       |
| volume list      | remove volume          | <kbd>d</kbd>                                       |
| volume list      | inspect volume         | <kbd>Enter</kbd>                                   |
| volume list      | refresh volume list    | <kbd>Ctrl</kbd> + <kbd>r</kbd>                     |
| volume list      | filter volume          | <kbd>/</kbd>                                       |
| network list     | inspect network        | <kbd>Enter</kbd>                                   |
| network list     | remove network         | <kbd>d</kbd>                                       |
| network list     | filter network         | <kbd>/</kbd>                                       |
| pull image       | pull image             | <kbd>Enter</kbd>                                   |
| pull image       | close panel            | <kbd>Esc</kbd>                                     |
| create container | next input box         | <kbd>Tab</kbd>                                     |
| create container | previous input box     | <kbd>Shift</kbd> +  <kbd>Tab</kbd>                 |
| detail           | cursor dwon            | <kbd>j</kbd>                                       |
| detail           | cursor up              | <kbd>k</kbd>                                       |
| detail           | next page              | <kbd>Ctrl</kbd> / <kbd>f</kbd>                     |
| detail           | previous page          | <kbd>Ctrl</kbd> / <kbd>b</kbd>                     |
| search images    | search image           | <kbd>Enter</kbd>                                   |
| search images    | close panel            | <kbd>Esc</kbd>                                     |
| search result    | next image             | <kbd>j</kbd>                                       |
| search result    | previous image         | <kbd>k</kbd>                                       |
| search result    | pull image             | <kbd>Enter</kbd>                                   |
| search result    | close panel            | <kbd>q</kbd>                                       |
| create volume    | close panel            | <kbd>Esc</kbd>                                     |
| create volume    | next input box         | <kbd>Tab</kbd>                                     |
| create volume    | previous input box     | <kbd>Shift</kbd> +  <kbd>Tab</kbd>                 |

## How to use
For details of the input panel please refer to [wiki](https://github.com/skanehira/docui/blob/master/wiki.md)

## Alternatives
- [lazydocker](https://github.com/jesseduffield/lazydocker)
A simple terminal UI for both docker and docker-compose, written in Go with the gocui library.
- [docker.vim](https://github.com/skanehira/docker.vim)
Manage docker containers and images in Vim
- See [Awesome Docker list](https://github.com/veggiemonk/awesome-docker/blob/master/README.md#terminal) for similar tools to work with Docker.
