FROM ubuntu:14.04.2

RUN apt-get update && apt-get install -y curl build-essential git-core
RUN curl -sSL https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz | tar -C /usr/local -xz
ENV PATH /usr/local/go/bin:$PATH
ENV GOPATH /go
ENV CGO_ENABLED 0

RUN mkdir $GOPATH && cd $GOPATH && \
	git clone https://github.com/fsouza/go-dockerclient.git && \
	git clone https://github.com/docker/docker.git && \
	cd go-dockerclient && git checkout a48995f21b2b00e5fc && cd .. && \
	mkdir -p src/github.com/fsouza && \
	mkdir -p src/github.com/docker && \
	ln -s $(pwd)/go-dockerclient src/github.com/fsouza && \
	ln -s $(pwd)/docker src/github.com/docker && \
	go get github.com/fsouza/go-dockerclient && \
	go get github.com/getgauge/mflag
COPY . /dlgrab/
RUN cd /dlgrab && make check && make binary

CMD /dlgrab/bin/dlgrab
