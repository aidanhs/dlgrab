FROM ubuntu:14.04

run apt-get update && apt-get install -y curl build-essential git-core

# Install Go (this is copied from the docker Dockerfile)
run curl -s https://go.googlecode.com/files/go1.1.1.linux-amd64.tar.gz | tar -v -C /usr/local -xz
env PATH  /usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/bin:/usr/sbin:/bin:/sbin
env GOPATH  /go
env CGO_ENABLED 0

add . /docker-registry.git/contrib/golang_impl
run cd /docker-registry.git/contrib/golang_impl && make && cp bin/docker-registry /usr/local/bin/

expose 5000
cmd /usr/local/bin/docker-registry
