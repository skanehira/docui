# docui - GUI Docker Client With Go

## About docui
docui is gui docker client.  
docui can do thises.

- image
    - search/pull/remove
    - save/import/load
    - inspect

- container
    - create/remove
    - start/stop
    - export/commit
    - inspect/rename

- volume
    - create/remove/prune
    - inspect

## Required Tools
- Go Ver.1.11
- Docker Engine Ver.18.06.1-ce
- Git

## Installation
If yo not install go and set GOPATH/GOBIN,  
you must install and set env before install docui.

```
$ mkdir $GOPATH/src
$ cd $GOPATH/src
$ git clone https://github.com/skanehira/docui
$ cd docui
$ go install
```

## Update
```
$ cd docui
$ go install
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
| all              | close panel            | <kbd>Esc</kbd>                  |
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
| image list       | refresh image          | <kbd>Ctrl</kbd> + <kbd>r</kbd>  |
| container list   | inspect container      | <kbd>Enter</kbd> / <kbd>o</kbd> |
| container list   | delete container       | <kbd>d</kbd>                    |
| container list   | next container         | <kbd>j</kbd>                    |
| container list   | previous container     | <kbd>k</kbd>                    |
| container list   | start container        | <kbd>u</kbd>                    |
| container list   | stop container         | <kbd>s</kbd>                    |
| container list   | export container       | <kbd>e</kbd>                    |
| container list   | commit container       | <kbd>c</kbd>                    |
| container list   | rename container       | <kbd>r</kbd>                    |
| container list   | refresh container      | <kbd>Ctrl</kbd> + <kbd>r</kbd>  |
| volume list      | create volume          | <kbd>c</kbd>                    |
| volume list      | remove volume          | <kbd>d</kbd>                    |
| volume list      | prune volume           | <kbd>p</kbd>                    |
| volume list      | inspect volume         | <kbd>Enter</kbd> / <kbd>o</kbd> |
| volume list      | inspect volume         | <kbd>Ctrl</kbd> + <kbd>r</kbd>  |
| pull image       | pull image             | <kbd>Enter</kbd>                |
| pull image       | close panel            | <kbd>Enter</kbd>                |
| create container | next input box         | <kbd>Ctrl</kbd> + <kbd>j</kbd>  |
| create container | previous input box     | <kbd>Ctrl</kbd> + <kbd>k</kbd>  |
| create container | close panel            | <kbd>Enter</kbd>                |
| create container | create container       | <kbd>Enter</kbd>                |
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
| create volume    | create volume          | <kbd>Enter</kbd>                |
| create volume    | close panel            | <kbd>Esc</kbd>                  |
| create volume    | next input box         | <kbd>Ctrl</kbd> + <kbd>j</kbd>  |
| create volume    | previous input box     | <kbd>Ctrl</kbd> + <kbd>k</kbd>  |


## How to use
For details of the input panel please refer to [wiki](https://github.com/skanehira/docui/blob/master/wiki.md)

## Screenshots

![](https://github.com/skanehira/docui/blob/images/images/s1.png)
![](https://github.com/skanehira/docui/blob/images/images/s2.png)
![](https://github.com/skanehira/docui/blob/images/images/s3.png)
![](https://github.com/skanehira/docui/blob/images/images/s4.png)
