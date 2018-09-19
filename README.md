# docui - GUI Docker Client With Go

## About docui
docui is gui docker client.  
docui can do thises.

- image
    - pull/remove
    - save/import/load
    - inspect

- container
    - create/remove
    - start/stop
    - export/commit
    - inspect

## Installation  
If yo not install go and set GOPATH/GOBIN,  
you must install and set env before install docui.


```
$ go get github.com/skanehira/docui
$ docui
```

## Update

```
$ go get -u github.com/skanehira/docui
```

## How to use
| panel            | operation           | key                            |
|------------------|---------------------|--------------------------------|
| all              | change panel        | <kbd>Tab</kbd>                 |
| all              | quit                | <kbd>Ctrl</kbd> + <kbd>q</kbd> |
| all              | quit                | <kbd>q</kbd>                   |
| image list       | pull image          | <kbd>p</kbd>                   |
| image list       | remove image        | <kbd>d</kbd>                   |
| image list       | create container    | <kbd>c</kbd>                   |
| image list       | display detail      | <kbd>Enter</kbd>               |
| image list       | display detail      | <kbd>o</kbd>                   |
| image list       | save image          | <kbd>s</kbd>                   |
| image list       | import image        | <kbd>i</kbd>                   |
| image list       | load image          | <kbd>Ctrl</kbd> + <kbd>l</kbd> |
| image list       | next image          | <kbd>j</kbd>                   |
| image list       | previous image      | <kbd>k</kbd>                   |
| container list   | display detail      | <kbd>Enter</kbd>               |
| container list   | display detail      | <kbd>o</kbd>                   |
| container list   | delete container    | <kbd>d</kbd>                   |
| container list   | next container      | <kbd>j</kbd>                   |
| container list   | previous container  | <kbd>k</kbd>                   |
| container list   | start container     | <kbd>u</kbd>                   |
| container list   | stop container      | <kbd>s</kbd>                   |
| container list   | export container    | <kbd>e</kbd>                   |
| container list   | commit container    | <kbd>Ctrl</kbd> + <kbd>c</kbd> |
| pull image       | do pull image       | <kbd>Enter</kbd>               |
| pull image       | close panel         | <kbd>Ctrl</kbd> + <kbd>w</kbd> |
| create container | next input box      | <kbd>Ctrl</kbd> + <kbd>j</kbd> |
| create container | previous input box  | <kbd>Ctrl</kbd> + <kbd>k</kbd> |
| create container | close panel         | <kbd>Ctrl</kbd> + <kbd>w</kbd> |
| create container | do create container | <kbd>Enter</kbd>               |
| detail           | cursor dwon         | <kbd>j</kbd>                   |
| detail           | cursor up           | <kbd>k</kbd>                   |
| detail           | page dwon           | <kbd>d</kbd>                   |
| detail           | page up             | <kbd>u</kbd>                   |


## Screenshots

![](https://github.com/skanehira/docui/blob/images/images/s1.png)
![](https://github.com/skanehira/docui/blob/images/images/s2.png)
![](https://github.com/skanehira/docui/blob/images/images/s3.png)
![](https://github.com/skanehira/docui/blob/images/images/s4.png)
