# About docui
docui is a simple TUI Client for docker.  
Supported OS is Linux/Mac only.  

Also, although it supports UNIX domain socket, TCP, http/https.

# Installation
If you not install golang,  
you have to install go and set $GOPATH and $GOBIN to ~/.bashrc.

## 1. Install go

### Mac
```sh
brew intall golang
```

### Linux
```sh
yum install golang
```

### Add ~/.bashrc
```sh
# add thises to ~/.bashrc
export GOPATH=/to/your/path
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
```

### Reload ~/.bashrc
```sh
resource ~/.bashrc
```

## 2. Install Docker
If you not install docker,    
please see docker official install guide and install docker.

https://www.docker.com/get-started  

## 3. Install Git
### Mac
```sh
brew install git
```

### Linux
```sh
yum install git
```

## 3. Install docui

This project uses [Go Modules](https://github.com/golang/go/wiki/Modules).  
Please use Go version 1.11.4 higher.

Use go get or git clone:

```sh
$ go get -d github.com/skanehira/docui
$ cd $GOPATH/src/github.com/skanehira/docui
$ GO111MODULE=on go install
```

```sh
$ git clone https://github.com/skanehira/docui.git
$ cd docui/
$ GO111MODULE=on go install
```

Make sure your PATH includes the $GOPATH/bin directory so your commands can be easily used:

```sh
export PATH=$PATH:$GOPATH/bin
```

## 4. Update docui

Use git pull:

```sh
$ git pull
$ GO111MODULE=on go install
```


## 5. Use on Docker
If you want to use docui on docker.

```sh
$ docker run --rm -itv /var/run/docker.sock:/var/run/docker.sock skanehira/docui
```

## 6. Build Docker Image
If you want to customize image.

```sh
$ make docker-build
```

# How to use
Refer to the [keybinding](https://github.com/skanehira/docui#Keybindings) for panel operation.  
I will explain the items of each panel here.

## pull image panel

- Name  
Please enter the docker image name you want to pull.  
If you want to specify the version, please input as below.

```
mysql:5.7
```

## search images panel

Please enter the image name on the Docker Hub you want to search.  
This operation works like `docker search`

## save image panel

Please enter the file path to save the selected image.  
It must be absolute path or relative path.

## import image panel

Please enter the path of the image you want to import.  
It must be absolute path or relative path.

## load image panel

Please enter the path of the image you want to load.   
It must be absolute path or relative path.

## create container panel
- Name  
Container name.

- HostPort  
Port of container to be mapping.

- Port  
Port of the host OS to be mapped.

- VolumeType  
Specify VolumeType bind or volume.

- HostVolume  
If VolumeType is bind, path of the host OS that you want to mount.
It's must be absolute path.
It's similar docker command `docker -v /to/host/path:/to/container/path`.

If VolumeType is volume, specify the docker volume.
It's similar docker command `docker -v docekr/volume:/to/container/path`.

- Volume  
Path of the container that you want to mount.
It must be absolute path.

- Image  
Selected image id.

- Attach  
If you want to attach container, please Enter.

- User  
If you want to attach container, please input user name.

- Env  
The environment variable setting value can be defined by variables like `$PATH`.
In that case, we will obtain the value from the OS environment variable.
If you want to add multiple environment variables, please input as below.

```
GOPATH=~/go,GOBIN=~/go/bin,PATH=$PATH
```

- Cmd
If you want to add command arguments,  
please input as below.

```
/bin/bash,hello
```

## export container panel
Please enter the file path to save the selected container.  
It must be absolute path or relative path.

## commit container panel
- Container  
Selected container id.  

- Repository  
Please enter the image name of the committed container.  

- Tag  
If tag is empty it will be latest.

## create volume panel
- Name  
Specify volume name.

- Driver  
Specify volume driver name.

- Labels  
Set metadata for a volume.  
If you want to specify multiple labels, please enter as below.  

```
OS=Linux TYPE=nfs
```

- Options  
Set driver specific options.  
If you want to specify multiple options, please enter as below.  

```
type=nfs o=addr=192.168.1.1,rw device=:/path/to/dir
```

## Configuration

### Command-Line Options

Support custom endpoint:

```sh
$ docui -h
Usage of docui:
  -api string
        api version (default "1.39")
  -ca string
        ca.pem file path
  -cert string
        cert.pem file path
  -endpoint string
        Docker endpoint (default "unix:///var/run/docker.sock")
  -key string
        key.pem file path
```

Or set environment variable:

- `DOCKER_HOST`
- `DOCKER_TLS_VERIFY`
- `DOCKER_CERT_PATH`

These environment variables take precedence over command-line options.
