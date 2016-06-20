FROM debian:jessie

RUN apt-get update && apt-get install --no-install-recommends -y \
    ca-certificates \
    curl \
    mercurial \
    git-core

RUN curl -s https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz | tar -v -C /usr/local -xz

ENV GOPATH /go
ENV GOROOT /usr/local/go
ENV PATH $PATH:/usr/local/go/bin:/go/bin




# go get to download all the deps
#RUN go get -u github.com/crosbymichael/skydock

ADD src /go/src
ADD plugins/ /plugins

#RUN cd /go/src/github.com/crosbymichael/skydock && go build


RUN cd /go/src/github.com/crosbymichael/skydock && go install . ./...

ENTRYPOINT ["/go/bin/skydock"]
