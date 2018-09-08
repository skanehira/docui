# docui - Docker Client Tool With Go

## About docui
docui is docker client cui tool.  
This tool can do thises.

- pull/remove image  
- display image detail
- create/remove container
- start/stop container
- display container detail

## Installation

```
$ go get "github.com/skanehira/docui"
$ docui
```

## How to use
| panel          | operation        | key      |
|----------------|------------------|----------|
| all            | change panel     | <kbd>Tab</kbd>      |
| all            | quit             | <kbd>Ctrl</kbd> + <kbd>q</kbd> |
| image list     | pull image       | <kbd>Ctrl</kbd> + <kbd>p</kbd> |
| image list     | remove image     | <kbd>Ctrl</kbd> + <kbd>d</kbd> |
| image list     | display detail      | <kbd>Enter</kbd>    |
| image list     | create container | <kbd>Ctrl</kbd> + <kbd>c</kbd> |
| image list | next image      | <kbd>Ctrl</kbd> + <kbd>j</kbd> |
| image list | previous image  | <kbd>Ctrl</kbd> + <kbd>k</kbd> |
| container list | display detail      | <kbd>Enter</kbd>    |
| container list | delete container | <kbd>Ctrl</kbd> + <kbd>d</kbd> |
| container list | next container      | <kbd>Ctrl</kbd> + <kbd>j</kbd> |
| container list | previous container  | <kbd>Ctrl</kbd> + <kbd>k</kbd> |
| container list | start container      | <kbd>Ctrl</kbd> + <kbd>u</kbd> |
| container list | stop container  | <kbd>Ctrl</kbd> + <kbd>s</kbd> |
| pull image       | do pull image       | <kbd>Enter</kbd>    |
| pull image       | close panel         | <kbd>Ctrl</kbd> + <kbd>w</kbd> |
| create container | next input box      | <kbd>Ctrl</kbd> + <kbd>j</kbd> |
| create container | previous input box  | <kbd>Ctrl</kbd> + <kbd>k</kbd> |
| create container | close panel         | <kbd>Ctrl</kbd> + <kbd>w</kbd> |
| create container | do create container | <kbd>Enter</kbd>    |

## Screenshots

![](https://github.com/skanehira/docui/blob/images/images/image_pull.png)
![](https://github.com/skanehira/docui/blob/images/images/image_detail.png)
![](https://github.com/skanehira/docui/blob/images/images/container_detail.png)
![](https://github.com/skanehira/docui/blob/images/images/container_create.png)
