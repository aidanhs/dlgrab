FROM ubuntu:14.04

RUN apt-get update && apt-get install -y curl build-essential git-core
RUN curl -sSL https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar -C /usr/local -xz
ENV PATH /usr/local/go/bin:$PATH
ENV GOPATH /go
ENV CGO_ENABLED 0

RUN go get github.com/aidanhs/go-dockerclient github.com/docker/docker/pkg/mflag
COPY . /dlgrab/
RUN cd /dlgrab && make check && make binary

CMD /dlgrab/bin/dlgrab
