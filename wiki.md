# About docui
docui is a simple GUI tool for docker running on terminal.  
Supported OS is Linux / Mac only.  

Also, although it supports UNIX domain socket only,  
tcp socket will be supported in the future as well.  

# Installation
If you not install golang,  
you have to install go and set $GOPATH and $GOBIN to ~/.bashrc.

## 1. Install go

### Mac
```
brew intall golang
```

### Linux
```
yum install golang
```

### Add ~/.bashrc
```
# add thises to ~/.bashrc
export GOPATH=/to/your/path
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
```

### Reload ~/.bashrc
```
resource ~/.bashrc
```

## 2. Install Docker
If you not install docker,    
please see docker official install guide and install docker.

https://www.docker.com/get-started  

## 3. Install Git
### Mac
```
brew install git
```

### Linux
```
yum install gitt
```

## 3. Install/Update docui
```
$ git clone https://github.com/skanehira/docui
$ cd docui
$ go install
```

# How to use
Refer to the [keybinding](https://github.com/skanehira/docui#Keybindings) for panel operation.  
I will explain the items of each panel here.

## pull imagee panel
![](https://github.com/skanehira/docui/blob/images/images/image_pull.png)

- Name  
Please enter the docker image name you want to pull.  
If you want to specify the version, please input as below.

```
mysql:5.7
```

## search images panel
![](https://github.com/skanehira/docui/blob/images/images/image_search.png)

Please enter the image name on the Docker Hub you want to search.  
This operation works like `docker search`

## save image panel
![](https://github.com/skanehira/docui/blob/images/images/image_save.png)

Please enter the file path to save the selected image.  
It must be absolute path or relative path.

## import image panel
![](https://github.com/skanehira/docui/blob/images/images/image_import.png)

Please enter the path of the image you want to import.  
It must be absolute path or relative path.

## load image panel
![](https://github.com/skanehira/docui/blob/images/images/image_load.png)

Please enter the path of the image you want to load.   
It must be absolute path or relative path.

## create container panel
![](https://github.com/skanehira/docui/blob/images/images/container_create.png)

- Name  
Container name.

- HostPort  
Port of container to be mapping.

- Port  
Port of the host OS to be mapped.

- HostVolume  
Path of the host OS that you want to mount.
It must be absolute path.

- Volume  
Path of the container that you want to mount.
It must be absolute path.

- Image  
Selected image id.

- Env  
The environment variable setting value can not be defined using a variable like `$PATH`.  
If you want to add multiple environment variables,  
Please input as below.

```
GOPATH=~/go,GOBIN=~/go/bin
```

- Cmd
If you want to add command arguments,  
please input as below.

```
/bin/bash,hello
```

## export container panel
![](https://github.com/skanehira/docui/blob/images/images/container_export.png)

Please enter the file path to save the selected container.  
It must be absolute path or relative path.

## commit container panel
![](https://github.com/skanehira/docui/blob/images/images/container_commit.png)

- Container  
Selected container id.  

- Repository  
Please enter the image name of the committed container.  

- Tag  
If tag is empty it will be latest.
