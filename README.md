# Virtual-IP

[![Build Status](https://travis-ci.org/darxkies/virtual-ip.svg?branch=master)](https://travis-ci.org/darxkies/virtual-ip)
[![Go Report Card](https://goreportcard.com/badge/github.com/darxkies/virtual-ip)](https://goreportcard.com/report/github.com/darxkies/virtual-ip)
[![GitHub release](https://img.shields.io/github/release/darxkies/virtual-ip.svg)](https://github.com/darxkies/virtual-ip/releases/latest)
![GitHub](https://img.shields.io/github/license/darxkies/virtual-ip.svg)


Virtual-IP can be used to share a Virtual/Floating IP address between many computers. The IP address is assigned to only one computer. If that computer goes down, the same IP address is then reassigned to another computer in the cluster.

## Features

- Self contained binary that can be downloaded from GitHub
- Docker Image on [Docker Hub](https://cloud.docker.com/repository/docker/darxkies/virtual-ip) 
- No external dependencies as in sevices
- Uses [Raft Consensus](https://raft.github.io/) for cluster communication

## Usage

Assuming that you have the following setup:

- three Linux servers with the IP addresses: 192.168.0.101, 192.168.0.102, 192.168.0.103)
- the three servers have the network interface eth1
- the Virtual-IP is 192.168.0.50
- the port 10000 for Raft on all three servers is free

### Binary

Download the binary from [release page](https://github.com/darxkies/virtual-ip/releases).

Then on each server run the following commands

Server1 (192.168.0.101):

```shell
sudo virtual-ip -id server1 -bind 192.168.0.101:10000 -peers server1=192.168.0.101:10000,server2=192.168.0.102:10000,server3=192.168.0.103:10000 -interface eth1 -virtual-ip 192.168.0.50
```

Server2 (192.168.0.102):

```shell
sudo virtual-ip -id server2 -bind 192.168.0.102:10000 -peers server1=192.168.0.101:10000,server2=192.168.0.102:10000,server3=192.168.0.103:10000 -interface eth1 -virtual-ip 192.168.0.50
```

Server3 (192.168.0.103):

```shell
sudo virtual-ip -id server3 -bind 192.168.0.102:10000 -peers server1=192.168.0.101:10000,server2=192.168.0.102:10000,server3=192.168.0.103:10000 -interface eth1 -virtual-ip 192.168.0.50
```

### Docker

Alternatively, the Docker Image can be used like this.

Server1 (192.168.0.101):

```shell
docker run -ti --rm --privileged --net=host darxkies/virtual-ip -id server1 -bind 192.168.0.101:10000 -peers server1=192.168.0.101:10000,server2=192.168.0.102:10000,server3=192.168.0.103:10000 -interface eth1 -virtual-ip 192.168.0.50
```

Server2 (192.168.0.102):

```shell
docker run -ti --rm --privileged --net=host darxkies/virtual-ip -id server2 -bind 192.168.0.102:10000 -peers server1=192.168.0.101:10000,server2=192.168.0.102:10000,server3=192.168.0.103:10000 -interface eth1 -virtual-ip 192.168.0.50
```

Server3 (192.168.0.103):

```shell
docker run -ti --rm --privileged --net=host darxkies/virtual-ip -id server3 -bind 192.168.0.102:10000 -peers server1=192.168.0.101:10000,server2=192.168.0.102:10000,server3=192.168.0.103:10000 -interface eth1 -virtual-ip 192.168.0.50
```

# Build

To build from source code you need make and Docker, then run the following commands:


```
git clone git@github.com:darxkies/virtual-ip.git
cd virtual-ip
make
```

The commands generate the binary virtual-ip in the current directory.


