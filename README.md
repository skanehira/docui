# docui - CUI Docker Client With Go

## About docui
docui is cui docker client.  
docui can do thises.

- image
    - search/pull/remove
    - save/import/load
    - inspect/filtering

- container
    - create/remove
    - start/stop
    - export/commit
    - inspect/rename/filtering

- volume
    - create/remove/prune
    - inspect/filtering

- network
    - remove
    - inspect/filtering

[![asciicast](https://asciinema.org/a/212109.svg)](https://asciinema.org/a/212109)

## Required Tools
- Go Ver.1.11
- Docker Engine Ver.18.06.1-ce
- Git

## Installation
If yo not install go and set GOPATH/GOBIN,  
you must install and set env before install docui.

```
$ mkdir $GOPATH/src
$ go get -u github.com/skanehira/docui
```

## Update
```
$ go get -u github.com/skanehira/docui
```

## Use on Docker
```
$ docker run --rm -itv /var/run/docker.sock:/var/run/docker.sock skanehira/docui
$ docui
```

## Build Docker Image
```
$ cd build
$ bash build.sh
```

## Keybindings
| panel            | operation              | key                             |
|------------------|------------------------|---------------------------------|
| all              | change panel           | <kbd>Tab</kbd>                  |
| all              | quit                   | <kbd>Ctrl</kbd> + <kbd>q</kbd>  |
| all              | quit                   | <kbd>q</kbd>                    |
| image list       | pull image             | <kbd>p</kbd>                    |
| image list       | search images          | <kbd>Ctrl</kbd> + <kbd>s</kbd>  |
| image list       | remove image           | <kbd>d</kbd>                    |
| image list       | create container       | <kbd>c</kbd>                    |
| image list       | inspect image          | <kbd>Enter</kbd> / <kbd>o</kbd> |
| image list       | save image             | <kbd>s</kbd>                    |
| image list       | import image           | <kbd>i</kbd>                    |
| image list       | load image             | <kbd>Ctrl</kbd> + <kbd>l</kbd>  |
| image list       | next image             | <kbd>j</kbd>                    |
| image list       | previous image         | <kbd>k</kbd>                    |
| image list       | remove dangling images | <kbd>Ctrl</kbd> + <kbd>d</kbd>  |
| image list       | refresh image list     | <kbd>Ctrl</kbd> + <kbd>r</kbd>  |
| image list       | filter image           | <kbd>f</kbd>                    |
| container list   | inspect container      | <kbd>Enter</kbd> / <kbd>o</kbd> |
| container list   | remove container       | <kbd>d</kbd>                    |
| container list   | next container         | <kbd>j</kbd>                    |
| container list   | previous container     | <kbd>k</kbd>                    |
| container list   | start container        | <kbd>u</kbd>                    |
| container list   | stop container         | <kbd>s</kbd>                    |
| container list   | export container       | <kbd>e</kbd>                    |
| container list   | commit container       | <kbd>c</kbd>                    |
| container list   | rename container       | <kbd>r</kbd>                    |
| container list   | refresh container list | <kbd>Ctrl</kbd> + <kbd>r</kbd>  |
| container list   | filter image           | <kbd>f</kbd>                    |
| volume list      | create volume          | <kbd>c</kbd>                    |
| volume list      | remove volume          | <kbd>d</kbd>                    |
| volume list      | prune volume           | <kbd>p</kbd>                    |
| volume list      | inspect volume         | <kbd>Enter</kbd> / <kbd>o</kbd> |
| volume list      | refresh volume list    | <kbd>Ctrl</kbd> + <kbd>r</kbd>  |
| volume list      | filter image           | <kbd>f</kbd>                    |
| network list     | inspect network        | <kbd>Enter</kbd> / <kbd>o</kbd> |
| network list     | remove network         | <kbd>d</kbd>                    |
| network list     | next netowrk           | <kbd>j</kbd>                    |
| network list     | previous network       | <kbd>k</kbd>                    |
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
| images           | next image             | <kbd>j</kbd>                    |
| images           | previous image         | <kbd>k</kbd>                    |
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
